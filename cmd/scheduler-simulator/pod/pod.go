package pod

import coreinformers "k8s.io/client-go/informers/core/v1"

type Service struct {
	PodInformer coreinformers.PodInformer
}
