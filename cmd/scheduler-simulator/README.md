# Web-based Kubernetes scheduler simulator

Hello world. Here is web-based Kubernetes scheduler simulator.
You can easily simulate kube-scheduler's behaviour with it.

### Run server

To run this simulator's server, you have to install Go and etcd.

You can install etcd with [kubernetes/kubernetes/hack/install-etcd.sh](https://github.com/kubernetes/kubernetes/blob/master/hack/install-etcd.sh).

```bash
make serve
```

It starts etcd and simulator-server.

### Run frontend

To run the frontend, please see [README.md](./web/README.md) on ./web dir.

## Contributing

For the first step, you have to prepare tools with make.

```bash
make tools
```

Also, you can run lint, format and test with make.

```bash
# test
make test
# lint
make lint
# format
make format
```

see [Makefile](Makefile) for more details.