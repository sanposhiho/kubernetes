package etcd

import (
	"context"

	"golang.org/x/xerrors"
	schedulerapi "k8s.io/kube-scheduler/config/v1beta1"
)

func (c *Client) PutSchedulerConfig(ctx context.Context, k string, cfg *schedulerapi.KubeSchedulerConfiguration) error {
	if err := c.put(ctx, k, cfg); err != nil {
		return xerrors.Errorf("put scheduler config: %w", err)
	}
	return nil
}

func (c *Client) GetSchedulerConfig(ctx context.Context, k string) (*schedulerapi.KubeSchedulerConfiguration, error) {
	cfg := schedulerapi.KubeSchedulerConfiguration{}
	if err := c.get(ctx, k, &cfg); err != nil {
		return nil, xerrors.Errorf("get scheduler config: %w", err)
	}
	return &cfg, nil
}
