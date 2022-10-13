package v1

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	kcptemplates "github.com/fgiloux/kcp-operator-sdk/plugins/manifests/v1/templates/config/kcp"
)

var _ plugin.CreateAPISubcommand = &createAPISubcommand{}

type createAPISubcommand struct {
	config   config.Config
	resource *resource.Resource
}

func (s *createAPISubcommand) InjectConfig(c config.Config) error {
	s.config = c

	return nil
}

func (s *createAPISubcommand) InjectResource(res *resource.Resource) error {
	s.resource = res

	return nil
}

func (s *createAPISubcommand) Scaffold(fs machinery.Filesystem) error {
	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(fs,
		// NOTE: kubebuilder's default permissions are only for root users
		machinery.WithDirectoryPermissions(0755),
		machinery.WithFilePermissions(0644),
		machinery.WithConfig(s.config),
		machinery.WithResource(s.resource),
	)

	// If the gvk is non-empty
	if s.resource.Group != "" || s.resource.Version != "" || s.resource.Kind != "" {
		if err := scaffold.Execute(
			&kcptemplates.APIExport{},
			&kcptemplates.PatchAPIExport{},
		); err != nil {
			return fmt.Errorf("error scaffolding manifests: %v", err)
		}
	}

	return nil
}
