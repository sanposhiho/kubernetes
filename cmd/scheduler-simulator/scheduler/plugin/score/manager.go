package score

import (
	"context"
	"errors"
	"math/rand"
	"sync"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"

	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// scorePluginManager has all scorePlugin.
type scorePluginManager struct {
	mu      *sync.Mutex
	plugins []*scorePlugin
}

func newScorePluginManager() *scorePluginManager {
	return &scorePluginManager{
		mu: new(sync.Mutex),
	}
}

func (s *scorePluginManager) registerPlugin(p *scorePlugin) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.plugins = append(s.plugins, p)
}

func (s *scorePluginManager) appendPluginToNodeScores(state *framework.CycleState, pluginName string, score framework.NodeScore) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := pluginToNodeScoresKey()
	d, err := state.Read(key)
	if err != nil {
		if !errors.Is(err, framework.ErrNotFound) {
			return xerrors.Errorf("read from cycle state: %w", err)
		}
		d = &scoringdata{scores: map[string]framework.NodeScoreList{}}
	}
	l, ok := d.(*scoringdata)
	if !ok {
		return xerrors.New("invalid data type")
	}
	l.scores[pluginName] = append(l.scores[pluginName], score)

	state.Write(key, l)
	return nil
}

func (s *scorePluginManager) getSelectedNode(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (string, *framework.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := pluginToNodeScoresKey()
	d, err := state.Read(key)
	if err != nil {
		if !errors.Is(err, framework.ErrNotFound) {
			return "", framework.AsStatus(xerrors.Errorf("read from cycle state: %w", err))
		}
		d = &scoringdata{}
	}
	data, ok := d.(*scoringdata)
	if !ok {
		return "", framework.AsStatus(xerrors.New("invalid data type"))
	}
	scores := data.scores

	// try to get from cache
	if data.selectedNode != "" {
		return data.selectedNode, nil
	}

	// run normalized score
	status := s.runAllNormalizedScore(ctx, state, pod, scores)
	if !status.IsSuccess() {
		return "", status
	}

	nodes := make([]string, 0, len(scores[s.plugins[0].Name()]))
	for i := range scores[s.plugins[0].Name()] {
		nodes = append(nodes, scores[s.plugins[0].Name()][i].Name)
	}

	// Summarize all scores.
	result := make(framework.NodeScoreList, 0, len(nodes))

	for i := range nodes {
		result = append(result, framework.NodeScore{Name: nodes[i], Score: 0})
		for j := range scores {
			result[i].Score += scores[j][i].Score
		}
	}

	finalnode := selectNode(result)

	// store result on selectedNode field
	data.selectedNode = finalnode
	state.Write(key, data)

	return finalnode, nil
}

func (s *scorePluginManager) runAllNormalizedScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, pluginToNodeScores framework.PluginToNodeScores) *framework.Status {
	wg := &sync.WaitGroup{}
	var returnStatus *framework.Status
	for _, p := range s.plugins {
		p := p
		wg.Add(1)
		go func() {
			s := p.RunNormalizeScore(ctx, state, pod, pluginToNodeScores[p.Name()])
			if !s.IsSuccess() {
				returnStatus = s
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// If one or more NormalizeScore are non-success, this func returns non success status.
	return returnStatus
}

func selectNode(nodeScoreList framework.NodeScoreList) string {
	if len(nodeScoreList) == 0 {
		return ""
	}

	maxScore := nodeScoreList[0].Score
	selected := nodeScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range nodeScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selected = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selected = ns.Name
			}
		}
	}
	return selected
}

func pluginToNodeScoresKey() framework.StateKey {
	return "simulator/scoringdata"
}

type scoringdata struct {
	scores       framework.PluginToNodeScores
	selectedNode string
}

func (n *scoringdata) Clone() framework.StateData {
	return n
}
