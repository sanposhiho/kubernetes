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

package testing

import (
	"fmt"

	fuzz "github.com/google/gofuzz"

	admissionregistrationfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/admissionregistration/fuzzer"
	"github.com/sanposhiho/kubernetes/pkg/apis/apps"
	appsfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/apps/fuzzer"
	autoscalingfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/autoscaling/fuzzer"
	batchfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/batch/fuzzer"
	certificatesfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/certificates/fuzzer"
	api "github.com/sanposhiho/kubernetes/pkg/apis/core"
	corefuzzer "github.com/sanposhiho/kubernetes/pkg/apis/core/fuzzer"
	discoveryfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/discovery/fuzzer"
	extensionsfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/extensions/fuzzer"
	flowcontrolfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/flowcontrol/fuzzer"
	networkingfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/networking/fuzzer"
	policyfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/policy/fuzzer"
	rbacfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/rbac/fuzzer"
	resourcefuzzer "github.com/sanposhiho/kubernetes/pkg/apis/resource/fuzzer"
	schedulingfuzzer "github.com/sanposhiho/kubernetes/pkg/apis/scheduling/fuzzer"
	storagefuzzer "github.com/sanposhiho/kubernetes/pkg/apis/storage/fuzzer"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apitesting "k8s.io/apimachinery/pkg/api/apitesting"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
)

// overrideGenericFuncs override some generic fuzzer funcs from k8s.io/apiserver in order to have more realistic
// values in a Kubernetes context.
func overrideGenericFuncs(codecs runtimeserializer.CodecFactory) []interface{} {
	return []interface{}{
		func(j *runtime.Object, c fuzz.Continue) {
			// TODO: uncomment when round trip starts from a versioned object
			if true { //c.RandBool() {
				*j = &runtime.Unknown{
					// We do not set TypeMeta here because it is not carried through a round trip
					Raw:         []byte(`{"apiVersion":"unknown.group/unknown","kind":"Something","someKey":"someValue"}`),
					ContentType: runtime.ContentTypeJSON,
				}
			} else {
				types := []runtime.Object{&api.Pod{}, &api.ReplicationController{}}
				t := types[c.Rand.Intn(len(types))]
				c.Fuzz(t)
				*j = t
			}
		},
		func(r *runtime.RawExtension, c fuzz.Continue) {
			// Pick an arbitrary type and fuzz it
			types := []runtime.Object{&api.Pod{}, &apps.Deployment{}, &api.Service{}}
			obj := types[c.Rand.Intn(len(types))]
			c.Fuzz(obj)

			var codec runtime.Codec
			switch obj.(type) {
			case *apps.Deployment:
				codec = apitesting.TestCodec(codecs, appsv1.SchemeGroupVersion)
			default:
				codec = apitesting.TestCodec(codecs, v1.SchemeGroupVersion)
			}

			// Convert the object to raw bytes
			bytes, err := runtime.Encode(codec, obj)
			if err != nil {
				panic(fmt.Sprintf("Failed to encode object: %v", err))
			}

			// Set the bytes field on the RawExtension
			r.Raw = bytes
		},
	}
}

// FuzzerFuncs is a list of fuzzer functions
var FuzzerFuncs = fuzzer.MergeFuzzerFuncs(
	overrideGenericFuncs,
	corefuzzer.Funcs,
	extensionsfuzzer.Funcs,
	appsfuzzer.Funcs,
	batchfuzzer.Funcs,
	autoscalingfuzzer.Funcs,
	rbacfuzzer.Funcs,
	policyfuzzer.Funcs,
	resourcefuzzer.Funcs,
	certificatesfuzzer.Funcs,
	admissionregistrationfuzzer.Funcs,
	storagefuzzer.Funcs,
	networkingfuzzer.Funcs,
	metafuzzer.Funcs,
	schedulingfuzzer.Funcs,
	discoveryfuzzer.Funcs,
	flowcontrolfuzzer.Funcs,
)
