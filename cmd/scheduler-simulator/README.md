# Web-based Kubernetes scheduler simulator

Hello world. Here is web-based Kubernetes scheduler simulator.
You can easily simulate kube-scheduler's behaviour with it.

## Run simulator

We have docker image and [docker-compose.yml](./docker-compose.yml) to use the simulator easily.

You can use it with the below cmd.

```bash
make docker_build_and_up
```

## Run simulator without Docker

You have to run frontend, server and etcd.

### Run simulator server and etcd

To run this simulator's server, you have to install Go and etcd.

You can install etcd with [kubernetes/kubernetes/hack/install-etcd.sh](https://github.com/kubernetes/kubernetes/blob/master/hack/install-etcd.sh).

```bash
make start
```

It starts etcd and simulator-server locally.

### Run simulator frontend

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
For the frontend, please see [README.md](./web/README.md) on ./web dir.
