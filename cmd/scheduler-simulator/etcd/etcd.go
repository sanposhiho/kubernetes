package etcd

import (
	"context"
	"encoding/json"
	"time"

	"go.etcd.io/etcd/clientv3"
	"golang.org/x/xerrors"
	"k8s.io/klog/v2"

	simulatorcfg "k8s.io/kubernetes/cmd/scheduler-simulator/config"
	"k8s.io/kubernetes/cmd/scheduler-simulator/errors"
)

type Client struct {
	Endpoints   []string
	DialTimeout time.Duration
	c           *clientv3.Client
}

const (
	dialTimeout = 5 * time.Second
)

func NewClient(cfg *simulatorcfg.Config) (
	*Client,
	func(), // function to close client
	error,
) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{cfg.EtcdURL},
		DialTimeout: dialTimeout,
	})
	// TODO: error handle
	// TODO: close
	if err != nil {
		return nil, nil, xerrors.Errorf("create new v3 client for etcd: %w", err)
	}

	closefn := func() {
		if err := cli.Close(); err != nil {
			klog.Errorf("failed to close v3 client for etcd: %w", err)
		}
	}

	return &Client{
		c: cli,
	}, closefn, nil
}

func (c *Client) get(ctx context.Context, k string, v interface{}) error {
	d, err := c.c.Get(ctx, k)
	if err != nil {
		return xerrors.Errorf("get data to etcd: %w", err)
	}
	if d.Count == 0 {
		return xerrors.Errorf("get from etcd: %w", errors.ErrNotFound)
	}

	if err := json.Unmarshal(d.Kvs[0].Value, v); err != nil {
		return xerrors.Errorf("unmarshal json: %w", err)
	}

	return nil
}

func (c *Client) put(ctx context.Context, k string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return xerrors.Errorf("encode to json: %w", err)
	}
	_, err = c.c.Put(ctx, k, string(data))
	if err != nil {
		return xerrors.Errorf("put data to etcd: %w", err)
	}

	return nil
}
