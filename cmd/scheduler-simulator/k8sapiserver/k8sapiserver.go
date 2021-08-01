package k8sapiserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"

	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/test/integration/framework"
)

// StartAPIServerOrDie starts API server, and it make panic when a error happen.
// TODO: change it not to use integration framework.
func StartAPIServerOrDie(etcdURL string) (*restclient.Config, func()) {
	h := &framework.APIServerHolder{Initialized: make(chan struct{})}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	}))

	// etcdOption for control plane
	etcdOptions := options.NewEtcdOptions(storagebackend.NewDefaultConfig(uuid.New().String(), nil))
	etcdOptions.StorageConfig.Transport.ServerList = []string{etcdURL}
	c := framework.NewIntegrationTestControlPlaneConfigWithOptions(&framework.MasterConfigOptions{
		EtcdOptions: etcdOptions,
	})
	c.GenericConfig.OpenAPIConfig = framework.DefaultOpenAPIConfig()

	// Note: This function die when a error happen.
	_, _, closeFn := framework.RunAnAPIServerUsingServer(c, s, h)

	cfg := &restclient.Config{
		Host:          s.URL,
		ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}},
		QPS:           5000.0,
		Burst:         5000,
	}

	shutdownFunc := func() {
		klog.Infof("destroying API server")
		closeFn()
		s.Close()
		klog.Infof("destroyed API server")
	}
	return cfg, shutdownFunc
}
