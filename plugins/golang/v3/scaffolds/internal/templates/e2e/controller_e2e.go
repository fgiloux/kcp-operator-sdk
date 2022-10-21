package e2e

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &E2ETest{}
var _ machinery.Inserter = &E2ETest{}

// E2ETest scaffolds the file that sets up the controller end-to-end tests
// nolint:maligned
type E2ETest struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin
}

// SetTemplateDefaults implements file.Template
func (f *E2ETest) SetTemplateDefaults() error {
	if f.Path == "" {
		if f.MultiGroup && f.Resource.Group != "" {
			f.Path = filepath.Join("test", "e2e", "%[group]", "controller_test.go")
		} else {
			f.Path = filepath.Join("test", "e2e", "controller_test.go")
		}
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	f.TemplateBody = fmt.Sprintf(e2eTestTemplate,
		machinery.NewMarkerFor(f.Path, importMarker),
	)

	return nil
}

const (
	importMarker = "imports"
)

// GetMarkers implements file.Inserter
func (f *E2ETest) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.Path, importMarker),
	}
}

const (
	apiImportCodeFragment = `%s "%s"`
)

// GetCodeFragments implements file.Inserter
func (f *E2ETest) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 1)

	// Generate import code fragments
	imports := make([]string, 0)
	if f.Resource.Path != "" {
		imports = append(imports, fmt.Sprintf(apiImportCodeFragment, f.Resource.ImportAlias(), f.Resource.Path))
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.Path, importMarker)] = imports
	}

	return fragments
}

const e2eTestTemplate = `{{ .Boilerplate }}

{{if and .MultiGroup .Resource.Group }}
package {{ .Resource.PackageName }}
{{else}}
package e2e
{{end}}

import (
	"flag"
	"math/rand"
	"testing"

	kcpclienthelper "github.com/kcp-dev/apimachinery/pkg/client"
	"github.com/kcp-dev/logicalcluster/v2"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
        tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"

	// corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client/config"

	%s
)

// The tests in this package expect to be called when:
// - kcp is running
// - a kind cluster is up and running
// - it is hosting a syncer, and the SyncTarget is ready to go
// - the controller-manager from this repo is deployed to kcp
// - that deployment is synced to the kind cluster
// - the deployment is rolled out & ready
//
// We can then check that the controllers defined here are working as expected.

var workspaceName string

func init() {
	rand.Seed(time.Now().Unix())
	flag.StringVar(&workspaceName, "workspace", "", "Workspace in which to run these tests.")
}

func parentWorkspace(t *testing.T) logicalcluster.Name {
        flag.Parse()
        if workspaceName == "" {
                t.Fatal("--workspace cannot be empty")
        }

        return logicalcluster.New(workspaceName)
}

func loadClusterConfig(t *testing.T, clusterName logicalcluster.Name) *rest.Config {
        t.Helper()
        restConfig, err := config.GetConfigWithContext("base")
        if err != nil {
                t.Fatalf("failed to load *rest.Config: %%v", err)
        }
        return rest.AddUserAgent(kcpclienthelper.SetCluster(rest.CopyConfig(restConfig), clusterName), t.Name())
}

func loadClient(t *testing.T, clusterName logicalcluster.Name) client.Client {
        t.Helper()
        scheme := runtime.NewScheme()
        if err := clientgoscheme.AddToScheme(scheme); err != nil {
                t.Fatalf("failed to add client go to scheme: %%v", err)
        }
        if err := tenancyv1alpha1.AddToScheme(scheme); err != nil {
                t.Fatalf("failed to add %%s to scheme: %%v", tenancyv1alpha1.SchemeGroupVersion, err)
        }
        if err := {{ .Resource.ImportAlias }}.AddToScheme(scheme); err != nil {
                t.Fatalf("failed to add %%s to scheme: %%v", {{ .Resource.ImportAlias }}.GroupVersion, err)
        }
        if err := apisv1alpha1.AddToScheme(scheme); err != nil {
                t.Fatalf("failed to add %%s to scheme: %%v", apisv1alpha1.SchemeGroupVersion, err)
        }
        tenancyClient, err := client.New(loadClusterConfig(t, clusterName), client.Options{Scheme: scheme})
        if err != nil {
                t.Fatalf("failed to create a client: %%v", err)
        }
        return tenancyClient
}

func createWorkspace(t *testing.T, clusterName logicalcluster.Name) client.Client {
        t.Helper()
        parent, ok := clusterName.Parent()
        if !ok {
                t.Fatalf("cluster %%s has no parent", clusterName)
        }
        c := loadClient(t, parent)
        t.Logf("creating workspace %%s", clusterName)
        if err := c.Create(context.TODO(), &tenancyv1alpha1.ClusterWorkspace{
                ObjectMeta: metav1.ObjectMeta{
                        Name: clusterName.Base(),
                },
                Spec: tenancyv1alpha1.ClusterWorkspaceSpec{
                        Type: tenancyv1alpha1.ClusterWorkspaceTypeReference{
                                Name: "universal",
                                Path: "root",
                        },
                },
        }); err != nil {
                t.Fatalf("failed to create workspace: %%s: %%v", clusterName, err)
        }

        t.Logf("waiting for workspace %%s to be ready", clusterName)
        var workspace tenancyv1alpha1.ClusterWorkspace
        if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
                fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: clusterName.Base()}, &workspace)
                if fetchErr != nil {
                        t.Logf("failed to get workspace %%s: %%v", clusterName, err)
                        return false, fetchErr
                }
                var reason string
                if actual, expected := workspace.Status.Phase, tenancyv1alpha1.ClusterWorkspacePhaseReady; actual != expected {
                        reason = fmt.Sprintf("phase is %%s, not %%s", actual, expected)
                        t.Logf("not done waiting for workspace %%s to be ready: %%s", clusterName, reason)
                }
                return reason == "", nil
        }); err != nil {
                t.Fatalf("workspace %%s never ready: %%v", clusterName, err)
        }

        return createAPIBinding(t, clusterName)
}

func createAPIBinding(t *testing.T, workspaceCluster logicalcluster.Name) client.Client {
        c := loadClient(t, workspaceCluster)
        apiName := "{{ .ProjectName }}-{{ .ProjectName }}.{{ .Domain }}"
        t.Logf("creating APIBinding %%s|%%s", workspaceCluster, apiName)
        if err := c.Create(context.TODO(), &apisv1alpha1.APIBinding{
                ObjectMeta: metav1.ObjectMeta{
                        Name: apiName,
                },
                Spec: apisv1alpha1.APIBindingSpec{
                        Reference: apisv1alpha1.ExportReference{
                                Workspace: &apisv1alpha1.WorkspaceExportReference{
                                        Path:       parentWorkspace(t).String(),
                                        ExportName: apiName,
                                },
                        },
                        // TODO(user): PermissionClaims need to be configured for the desired resources
                        // Example:
                        // PermissionClaims: []apisv1alpha1.AcceptablePermissionClaim{
                        //      {
                        //              PermissionClaim: apisv1alpha1.PermissionClaim{
                        //                      GroupResource: apisv1alpha1.GroupResource{Resource: "configmaps"},
                        //              },
                        //              State: apisv1alpha1.ClaimAccepted,
                        //      },
                        // },
                },
        }); err != nil {
                t.Fatalf("could not create APIBinding %%s|%%s: %%v", workspaceCluster, apiName, err)
        }

        t.Logf("waiting for APIBinding %%s|%%s to be bound", workspaceCluster, apiName)
        var apiBinding apisv1alpha1.APIBinding
        if err := wait.PollImmediate(100*time.Millisecond, wait.ForeverTestTimeout, func() (done bool, err error) {
                fetchErr := c.Get(context.TODO(), client.ObjectKey{Name: apiName}, &apiBinding)
                if fetchErr != nil {
                        t.Logf("failed to get APIBinding %%s|%%s: %%v", workspaceCluster, apiName, err)
                        return false, fetchErr
                }
                var reason string
                if !conditions.IsTrue(&apiBinding, apisv1alpha1.InitialBindingCompleted) {
                        condition := conditions.Get(&apiBinding, apisv1alpha1.InitialBindingCompleted)
                        if condition != nil {
                                reason = fmt.Sprintf("%%s: %%s", condition.Reason, condition.Message)
                        } else {
                                reason = "no condition present"
                        }
                        t.Logf("not done waiting for APIBinding %%s|%%s to be bound: %%s", workspaceCluster, apiName, reason)
                }
                return conditions.IsTrue(&apiBinding, apisv1alpha1.InitialBindingCompleted), nil
        }); err != nil {
                t.Fatalf("APIBinding %%s|%%s never bound: %%v", workspaceCluster, apiName, err)
        }

        return c
}

const characters = "abcdefghijklmnopqrstuvwxyz"

func randomName() string {
        b := make([]byte, 10)
        for i := range b {
                b[i] = characters[rand.Intn(len(characters))]
        }
        return string(b)
}

// TestController verifies that the controller behavior works.
func TestController(t *testing.T) {
        t.Parallel()
        for i := 0; i < 3; i++ {
                t.Run(fmt.Sprintf("attempt-%%d", i), func(t *testing.T) {
                        t.Parallel()
                        workspaceCluster := parentWorkspace(t).Join(randomName())
                        c := createWorkspace(t, workspaceCluster)
			t.Logf("workspace client %v", c)

                        // TODO(user): Create resources and check that the desired reconciliation took place.
                        // Example: 
                        // namespaceName := randomName()
                        // t.Logf("creating namespace %%s|%%s", workspaceCluster, namespaceName)
                        // if err := c.Create(context.TODO(), &corev1.Namespace{
                        //     ObjectMeta: metav1.ObjectMeta{Name: namespaceName},}); err != nil {
                        //              t.Fatalf("failed to create a namespace: %%v", err)
                        // }
                        // if err := c.Create(context.TODO(), &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{
                        //     ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: fmt.Sprintf("resource-%%d", i)},
                        //     Spec: {{ .Resource.ImportAlias }}.{{ .Resource.Kind }}Spec{},
                        // }); err != nil {
                        //     t.Fatalf("failed to create {{ .Resource.Kind }}: %%v", err)
                        // }
                })
        }
}

`
