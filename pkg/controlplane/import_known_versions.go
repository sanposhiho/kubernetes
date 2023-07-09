/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controlplane

import (
	// These imports are the API groups the API server will support.
	_ "github.com/sanposhiho/kubernetes/pkg/apis/admission/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/admissionregistration/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/apiserverinternal/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/apps/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/authentication/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/authorization/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/autoscaling/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/batch/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/certificates/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/coordination/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/core/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/discovery/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/events/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/extensions/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/flowcontrol/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/imagepolicy/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/networking/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/node/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/policy/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/rbac/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/resource/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/scheduling/install"
	_ "github.com/sanposhiho/kubernetes/pkg/apis/storage/install"
)
