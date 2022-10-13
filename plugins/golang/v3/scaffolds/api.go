package scaffolds

import (
	"fmt"

	"github.com/spf13/afero"

	"github.com/fgiloux/kcp-operator-sdk/plugins/golang/v3/scaffolds/internal/templates"
	"github.com/fgiloux/kcp-operator-sdk/plugins/golang/v3/scaffolds/internal/templates/api"
	"github.com/fgiloux/kcp-operator-sdk/plugins/golang/v3/scaffolds/internal/templates/controllers"
	"github.com/fgiloux/kcp-operator-sdk/plugins/golang/v3/scaffolds/internal/templates/hack"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
)

var _ plugins.Scaffolder = &apiScaffolder{}

// apiScaffolder contains configuration for generating scaffolding for Go type
// representing the API and controller that implements the behavior for the API.
type apiScaffolder struct {
	config   config.Config
	resource resource.Resource

	// fs is the filesystem that will be used by the scaffolder
	fs machinery.Filesystem

	// force indicates whether to scaffold controller files even if it exists or not
	force bool
}

// NewAPIScaffolder returns a new Scaffolder for API/controller creation operations
func NewAPIScaffolder(config config.Config, res resource.Resource, force bool) plugins.Scaffolder {
	return &apiScaffolder{
		config:   config,
		resource: res,
		force:    force,
	}
}

// InjectFS implements cmdutil.Scaffolder
func (s *apiScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// Scaffold implements cmdutil.Scaffolder
func (s *apiScaffolder) Scaffold() error {
	fmt.Println("Writing scaffold for you to edit...")

	// Load the boilerplate
	boilerplate, err := afero.ReadFile(s.fs.FS, hack.DefaultBoilerplatePath)
	if err != nil {
		return fmt.Errorf("error scaffolding API/controller: unable to load boilerplate: %w", err)
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
		machinery.WithResource(&s.resource),
	)

	// Keep track of these values before the update
	doAPI := s.resource.HasAPI()
	doController := s.resource.HasController()

	if err := s.config.UpdateResource(s.resource); err != nil {
		return fmt.Errorf("error updating resource: %w", err)
	}

	if doAPI {
		if err := scaffold.Execute(
			&api.Types{Force: s.force},
			&api.Group{},
		); err != nil {
			return fmt.Errorf("error scaffolding APIs: %v", err)
		}
	}

	if doController {
		if err := scaffold.Execute(
			&controllers.SuiteTest{Force: s.force},
			&controllers.Controller{ControllerRuntimeVersion: ControllerRuntimeVersion, Force: s.force},
		); err != nil {
			return fmt.Errorf("error scaffolding controller: %v", err)
		}
	}

	if err := scaffold.Execute(
		&templates.MainUpdater{WireResource: doAPI, WireController: doController},
	); err != nil {
		return fmt.Errorf("error updating main.go: %v", err)
	}

	return nil
}
