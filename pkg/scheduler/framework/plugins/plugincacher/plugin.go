package plugincacher

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)


type cacherExtension interface {
	FlushCache(cacherHandle CacherHandle)
}

// PluginCacher actually isn't a real plugin, it's just a wrapper;
// it wraps one filter plugin and act as that plugin instead.
//
// What's "wraps" meaning?
// It implements the Filter interface, observe the request/response of a plugin, and cache the response.
// The response is cached with the information that it's for which Pod and for which Node.
type PluginCacher struct {
	plugin framework.CachedPlugin

	filterCache cache
}

type cache struct {
	mu sync.RWMutex

	// cache is the actual cache data.
	// keyed by a Node name + a Pod name + a Pod Namespace name. (We can use it as an unique key)
	cache map[string]framework.Status

	// nodeIndex stores which keys has the cache data related to a Node.
	// It's keyed by a Node name.
	nodeIndex map[string][]string
	// podIndex stores which keys has the cache data related to a Pod.
	// It's keyed by a Pod name + a Namespace name.
	podIndex map[string][]string

	// allowedPods are Pods allowed to be moved to activeQ.
	allowedPods sets.Set[string]
}

var _ framework.CacheHandle = cache 

func buildCacheKey(nodeName, podName, namespaceName string) string {
	return fmt.Sprintf("%s/%s/%s", nodeName, podName, namespaceName)
}

func extractPodName(key string) (string, error) {
	splits := strings.Split(key, "/")
	if len(splits) != 3 {
		return "", errors.New("invalid key. The key contains more than 3 `/`.")
	}
	return splits[0], nil
}

func (p *PluginCacher) HandleCache(event framework.ClusterEvent, involvedObj runtime.Object) {
	p.plugin.HandleCache(&p.filterCache, event, involvedObj)
}

func (p *PluginCacher) PreEnqueue(ctx context.Context, p *v1.Pod) *framework.Status {
}

func (p *PluginCacher) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
}

// FlushNodeCache clean up all cache related to a given Node.
func (c *cache) FlushNodeCache(nodename string) {
	keys, ok := c.nodeIndex[nodename]
	if !ok {
		return
	}

	for _, key := range keys {
		podName, err := extractPodName(key)
		if err != nil {
			// TODO: klog
		}
		c.allowedPods.Insert(podName)
		delete(c.cache, key)
	}

}

// FlushPodCache clean up all cache related to a given Pod.
func (c *cache) FlushPodCache(podname string) {
	keys, ok := c.podIndex[podname]
	if !ok {
		return
	}

	for _, key := range keys {
		delete(c.cache, key)
	}

	c.allowedPods.Insert(podname)
}

func New(p framework.Plugin) framework.Plugin {
	fp, ok := p.(framework.FilterPlugin)
	if !ok {
		panic("")
	}

}
