/*
Copyright 2014 The Kubernetes Authors.

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

package options

// This file exists to force the desired plugin implementations to be linked.
// This should probably be part of some configuration fed into the build for a
// given binary target.
import (
	"k8s.io/apiserver/pkg/admission/plugin/validatingadmissionpolicy"
	// Admission policies
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/admit"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/alwayspullimages"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/antiaffinity"
	certapproval "github.com/sanposhiho/kubernetes/plugin/pkg/admission/certificates/approval"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/certificates/ctbattest"
	certsigning "github.com/sanposhiho/kubernetes/plugin/pkg/admission/certificates/signing"
	certsubjectrestriction "github.com/sanposhiho/kubernetes/plugin/pkg/admission/certificates/subjectrestriction"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/defaulttolerationseconds"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/deny"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/eventratelimit"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/extendedresourcetoleration"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/gc"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/imagepolicy"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/limitranger"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/namespace/autoprovision"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/namespace/exists"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/network/defaultingressclass"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/network/denyserviceexternalips"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/noderestriction"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/nodetaint"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/podnodeselector"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/podtolerationrestriction"
	podpriority "github.com/sanposhiho/kubernetes/plugin/pkg/admission/priority"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/runtimeclass"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/security/podsecurity"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/securitycontext/scdeny"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/serviceaccount"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/storage/persistentvolume/label"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/storage/persistentvolume/resize"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/storage/storageclass/setdefault"
	"github.com/sanposhiho/kubernetes/plugin/pkg/admission/storage/storageobjectinuseprotection"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/plugin/namespace/lifecycle"
	"k8s.io/apiserver/pkg/admission/plugin/resourcequota"
	mutatingwebhook "k8s.io/apiserver/pkg/admission/plugin/webhook/mutating"
	validatingwebhook "k8s.io/apiserver/pkg/admission/plugin/webhook/validating"
)

// AllOrderedPlugins is the list of all the plugins in order.
var AllOrderedPlugins = []string{
	admit.PluginName,                        // AlwaysAdmit
	autoprovision.PluginName,                // NamespaceAutoProvision
	lifecycle.PluginName,                    // NamespaceLifecycle
	exists.PluginName,                       // NamespaceExists
	scdeny.PluginName,                       // SecurityContextDeny
	antiaffinity.PluginName,                 // LimitPodHardAntiAffinityTopology
	limitranger.PluginName,                  // LimitRanger
	serviceaccount.PluginName,               // ServiceAccount
	noderestriction.PluginName,              // NodeRestriction
	nodetaint.PluginName,                    // TaintNodesByCondition
	alwayspullimages.PluginName,             // AlwaysPullImages
	imagepolicy.PluginName,                  // ImagePolicyWebhook
	podsecurity.PluginName,                  // PodSecurity
	podnodeselector.PluginName,              // PodNodeSelector
	podpriority.PluginName,                  // Priority
	defaulttolerationseconds.PluginName,     // DefaultTolerationSeconds
	podtolerationrestriction.PluginName,     // PodTolerationRestriction
	eventratelimit.PluginName,               // EventRateLimit
	extendedresourcetoleration.PluginName,   // ExtendedResourceToleration
	label.PluginName,                        // PersistentVolumeLabel
	setdefault.PluginName,                   // DefaultStorageClass
	storageobjectinuseprotection.PluginName, // StorageObjectInUseProtection
	gc.PluginName,                           // OwnerReferencesPermissionEnforcement
	resize.PluginName,                       // PersistentVolumeClaimResize
	runtimeclass.PluginName,                 // RuntimeClass
	certapproval.PluginName,                 // CertificateApproval
	certsigning.PluginName,                  // CertificateSigning
	ctbattest.PluginName,                    // ClusterTrustBundleAttest
	certsubjectrestriction.PluginName,       // CertificateSubjectRestriction
	defaultingressclass.PluginName,          // DefaultIngressClass
	denyserviceexternalips.PluginName,       // DenyServiceExternalIPs

	// new admission plugins should generally be inserted above here
	// webhook, resourcequota, and deny plugins must go at the end

	mutatingwebhook.PluginName,           // MutatingAdmissionWebhook
	validatingadmissionpolicy.PluginName, // ValidatingAdmissionPolicy
	validatingwebhook.PluginName,         // ValidatingAdmissionWebhook
	resourcequota.PluginName,             // ResourceQuota
	deny.PluginName,                      // AlwaysDeny
}

// RegisterAllAdmissionPlugins registers all admission plugins.
// The order of registration is irrelevant, see AllOrderedPlugins for execution order.
func RegisterAllAdmissionPlugins(plugins *admission.Plugins) {
	admit.Register(plugins) // DEPRECATED as no real meaning
	alwayspullimages.Register(plugins)
	antiaffinity.Register(plugins)
	defaulttolerationseconds.Register(plugins)
	defaultingressclass.Register(plugins)
	denyserviceexternalips.Register(plugins)
	deny.Register(plugins) // DEPRECATED as no real meaning
	eventratelimit.Register(plugins)
	extendedresourcetoleration.Register(plugins)
	gc.Register(plugins)
	imagepolicy.Register(plugins)
	limitranger.Register(plugins)
	autoprovision.Register(plugins)
	exists.Register(plugins)
	noderestriction.Register(plugins)
	nodetaint.Register(plugins)
	label.Register(plugins) // DEPRECATED, future PVs should not rely on labels for zone topology
	podnodeselector.Register(plugins)
	podtolerationrestriction.Register(plugins)
	runtimeclass.Register(plugins)
	resourcequota.Register(plugins)
	podsecurity.Register(plugins)
	podpriority.Register(plugins)
	scdeny.Register(plugins)
	serviceaccount.Register(plugins)
	setdefault.Register(plugins)
	resize.Register(plugins)
	storageobjectinuseprotection.Register(plugins)
	certapproval.Register(plugins)
	certsigning.Register(plugins)
	ctbattest.Register(plugins)
	certsubjectrestriction.Register(plugins)
}

// DefaultOffAdmissionPlugins get admission plugins off by default for kube-apiserver.
func DefaultOffAdmissionPlugins() sets.String {
	defaultOnPlugins := sets.NewString(
		lifecycle.PluginName,                    // NamespaceLifecycle
		limitranger.PluginName,                  // LimitRanger
		serviceaccount.PluginName,               // ServiceAccount
		setdefault.PluginName,                   // DefaultStorageClass
		resize.PluginName,                       // PersistentVolumeClaimResize
		defaulttolerationseconds.PluginName,     // DefaultTolerationSeconds
		mutatingwebhook.PluginName,              // MutatingAdmissionWebhook
		validatingwebhook.PluginName,            // ValidatingAdmissionWebhook
		resourcequota.PluginName,                // ResourceQuota
		storageobjectinuseprotection.PluginName, // StorageObjectInUseProtection
		podpriority.PluginName,                  // Priority
		nodetaint.PluginName,                    // TaintNodesByCondition
		runtimeclass.PluginName,                 // RuntimeClass
		certapproval.PluginName,                 // CertificateApproval
		certsigning.PluginName,                  // CertificateSigning
		ctbattest.PluginName,                    // ClusterTrustBundleAttest
		certsubjectrestriction.PluginName,       // CertificateSubjectRestriction
		defaultingressclass.PluginName,          // DefaultIngressClass
		podsecurity.PluginName,                  // PodSecurity
		validatingadmissionpolicy.PluginName,    // ValidatingAdmissionPolicy, only active when feature gate ValidatingAdmissionPolicy is enabled
	)

	return sets.NewString(AllOrderedPlugins...).Difference(defaultOnPlugins)
}
