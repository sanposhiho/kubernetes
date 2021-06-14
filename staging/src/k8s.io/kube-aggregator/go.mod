// This is a generated file. Do not edit directly.

module k8s.io/kube-aggregator

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/emicklei/go-restful v2.9.5+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/json-iterator/go v1.1.10
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/code-generator v0.0.0
	k8s.io/component-base v0.0.0
	k8s.io/klog/v2 v2.9.0
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e
	k8s.io/utils v0.0.0-20210521133846-da695404a2bc
	sigs.k8s.io/structured-merge-diff/v4 v4.1.1
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.54.0
	github.com/gregjones/httpcache => github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/grpc-ecosystem/go-grpc-middleware => github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/grpc-ecosystem/grpc-gateway => github.com/grpc-ecosystem/grpc-gateway v1.9.5
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.5
	github.com/jonboulle/clockwork => github.com/jonboulle/clockwork v0.1.0
	github.com/nxadm/tail => github.com/nxadm/tail v1.4.4
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.7.0
	github.com/tmc/grpc-websocket-proxy => github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net => golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
	golang.org/x/sync => golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	honnef.co/go/tools => honnef.co/go/tools v0.0.1-2020.1.3
	k8s.io/api => ../api
	k8s.io/apimachinery => ../apimachinery
	k8s.io/apiserver => ../apiserver
	k8s.io/client-go => ../client-go
	k8s.io/code-generator => ../code-generator
	k8s.io/component-base => ../component-base
	k8s.io/kube-aggregator => ../kube-aggregator
)
