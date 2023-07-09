/*
Copyright 2021 The Kubernetes Authors.

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

package apps

import (
	"context"

	"github.com/sanposhiho/kubernetes/test/e2e/cloud/gcp/common"
	"github.com/sanposhiho/kubernetes/test/e2e/framework"
	e2epv "github.com/sanposhiho/kubernetes/test/e2e/framework/pv"
	"github.com/sanposhiho/kubernetes/test/e2e/upgrades"
	"github.com/sanposhiho/kubernetes/test/e2e/upgrades/apps"
	"github.com/sanposhiho/kubernetes/test/utils/junit"
	admissionapi "k8s.io/pod-security-admission/api"

	"github.com/onsi/ginkgo/v2"
)

var upgradeTests = []upgrades.Test{
	&apps.MySQLUpgradeTest{},
	&apps.EtcdUpgradeTest{},
	&apps.CassandraUpgradeTest{},
}

var _ = SIGDescribe("stateful Upgrade [Feature:StatefulUpgrade]", func() {
	f := framework.NewDefaultFramework("stateful-upgrade")
	f.NamespacePodSecurityLevel = admissionapi.LevelPrivileged
	testFrameworks := upgrades.CreateUpgradeFrameworks(upgradeTests)

	ginkgo.Describe("stateful upgrade", func() {
		ginkgo.It("should maintain a functioning cluster", func(ctx context.Context) {
			e2epv.SkipIfNoDefaultStorageClass(ctx, f.ClientSet)
			upgCtx, err := common.GetUpgradeContext(f.ClientSet.Discovery())
			framework.ExpectNoError(err)

			testSuite := &junit.TestSuite{Name: "Stateful upgrade"}
			statefulUpgradeTest := &junit.TestCase{Name: "[sig-apps] stateful-upgrade", Classname: "upgrade_tests"}
			testSuite.TestCases = append(testSuite.TestCases, statefulUpgradeTest)

			upgradeFunc := common.ClusterUpgradeFunc(f, upgCtx, statefulUpgradeTest, nil, nil)
			upgrades.RunUpgradeSuite(ctx, upgCtx, upgradeTests, testFrameworks, testSuite, upgrades.ClusterUpgrade, upgradeFunc)
		})
	})
})
