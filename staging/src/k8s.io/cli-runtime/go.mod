// This is a generated file. Do not edit directly.

module k8s.io/cli-runtime

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/google/uuid v1.1.2
	github.com/googleapis/gnostic v0.5.1
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/text v0.3.6
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e
	sigs.k8s.io/kustomize/api v0.8.10
	sigs.k8s.io/yaml v1.2.0
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
	k8s.io/api => ../api
	k8s.io/apimachinery => ../apimachinery
	k8s.io/cli-runtime => ../cli-runtime
	k8s.io/client-go => ../client-go
)
