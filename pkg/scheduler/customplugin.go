package scheduler

import (
	"plugin"

	"k8s.io/klog/v2"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

func loadPlugins(files map[string]string) frameworkruntime.Registry {
	registry := map[string]frameworkruntime.PluginFactory{}

	for k, f := range files {
		p, err := plugin.Open(f)
		if err != nil {
			klog.Fatalf("failed to open plugin file: %s, %v", f, err)
		}

		newfunc, err := p.Lookup("New")
		if err != nil {
			klog.Fatalf("New function is not found on custom plugin file: %v", err)
		}

		factory, ok := newfunc.(frameworkruntime.PluginFactory)
		if !ok {
			klog.Fatalf("New function is not frameworkruntime.PluginFactory")
		}

		registry[k] = factory
	}

	return registry
}
