package schedulerconfig

import (
	"context"

	"golang.org/x/xerrors"
	schedulerapi "k8s.io/kube-scheduler/config/v1beta1"
)

type Service struct {
	store Store
}

type Store interface {
	GetSchedulerConfig(ctx context.Context, k string) (*schedulerapi.KubeSchedulerConfiguration, error)
	PutSchedulerConfig(ctx context.Context, k string, cfg *schedulerapi.KubeSchedulerConfiguration) error
}

// NewSchedulerConfigService initializes Service.
func NewSchedulerConfigService(s Store) *Service {
	return &Service{
		store: s,
	}
}

func (s *Service) GetSchedulerConfig(ctx context.Context, simulatorID string) (*schedulerapi.KubeSchedulerConfiguration, error) {
	c, err := s.store.GetSchedulerConfig(ctx, simulatorID)
	if err != nil {
		return nil, xerrors.Errorf("get scheduler config from store: %w", err)
	}
	return c, nil
}

func (s *Service) PutSchedulerConfig(ctx context.Context, simulatorID string, cfg *schedulerapi.KubeSchedulerConfiguration) error {
	if err := s.store.PutSchedulerConfig(ctx, simulatorID, cfg); err != nil {
		return xerrors.Errorf("put scheduler config to store: %w", err)
	}
	return nil
}
