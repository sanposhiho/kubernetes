// This is a generated file. Do not edit directly.

module k8s.io/kubelet

go 1.16

require (
	github.com/gogo/protobuf v1.3.2
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/grpc v1.27.1
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/component-base v0.0.0
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.54.0
	github.com/gregjones/httpcache => github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/grpc-ecosystem/go-grpc-middleware => github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.9.5
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.5
	github.com/mattn/go-isatty => github.com/mattn/go-isatty v0.0.4
	github.com/nxadm/tail => github.com/nxadm/tail v1.4.4
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.7.0
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net => golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
	honnef.co/go/tools => honnef.co/go/tools v0.0.1-2020.1.3
	k8s.io/api => ../api
	k8s.io/apimachinery => ../apimachinery
	k8s.io/client-go => ../client-go
	k8s.io/component-base => ../component-base
	k8s.io/kubelet => ../kubelet
)
