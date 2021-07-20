# Web-based Kubernetes scheduler simulator

This project is now under developing... 

(issue: [web-based simulator for scheduler behaviour #99605](https://github.com/kubernetes/kubernetes/issues/99605))

## Contributing

For the first step, you have to prepare tools with make.

```sh
make tools
```

Also, you can run lint, format and test with make.

```sh
# test
make test
# lint
make lint
# format
make format
```

see [Makefile](Makefile) for more details.

### Run server

```sh
make serve
```

It starts etcd and simulator-server.
You have to install etcd with [../../hack/install-etcd.sh](../../hack/install-etcd.sh).

### Run frontend

For the frontend, please see [README.md](./web/README.md) on ./web dir.