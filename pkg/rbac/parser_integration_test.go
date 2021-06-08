package rbac_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	rbacv1 "k8s.io/api/rbac/v1"

	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/rbac"
	"sigs.k8s.io/yaml"
)

var _ = Describe("ClusterRole generated by the RBAC Generator", func() {
	// run this test multiple times to make sure the Rule order is stable.
	const stableTestCount = 5
	for i := 0; i < stableTestCount; i++ {
		It("should match the expected result", func() {
			By("switching into testdata to appease go modules")
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chdir("./testdata")).To(Succeed()) // go modules are directory-sensitive
			defer func() { Expect(os.Chdir(cwd)).To(Succeed()) }()

			By("loading the roots")
			pkgs, err := loader.LoadRoots(".")
			Expect(err).NotTo(HaveOccurred())

			By("registering RBAC rule marker")
			reg := &markers.Registry{}
			Expect(reg.Register(rbac.RuleDefinition)).To(Succeed())

			By("creating GenerationContext")
			ctx := &genall.GenerationContext{
				Collector: &markers.Collector{Registry: reg},
				Roots:     pkgs,
			}

			By("generating a ClusterRole")
			objs, err := rbac.GenerateRoles(ctx, "manager-role")
			Expect(err).NotTo(HaveOccurred())

			By("loading the desired YAML")
			expectedFile, err := ioutil.ReadFile("role.yaml")
			Expect(err).NotTo(HaveOccurred())

			By("parsing the desired YAML")
			for i, expectedRoleBytes := range bytes.Split(expectedFile, []byte("\n---\n"))[1:] {
				By(fmt.Sprintf("comparing the generated Role and expected Role (Pair %d)", i))
				obj := objs[i]
				switch obj := obj.(type) {
				case rbacv1.ClusterRole:
					var expectedClusterRole rbacv1.ClusterRole
					Expect(yaml.Unmarshal(expectedRoleBytes, &expectedClusterRole)).To(Succeed())
					Expect(obj).To(Equal(expectedClusterRole), "type not as expected, check pkg/rbac/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(obj, expectedClusterRole))
				default:
					var expectedRole rbacv1.Role
					Expect(yaml.Unmarshal(expectedRoleBytes, &expectedRole)).To(Succeed())
					Expect(obj).To(Equal(expectedRole), "type not as expected, check pkg/rbac/testdata/README.md for more details.\n\nDiff:\n\n%s", cmp.Diff(obj, expectedRole))
				}
			}

		})
	}
})
