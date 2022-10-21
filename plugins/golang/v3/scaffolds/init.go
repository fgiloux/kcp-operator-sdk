package scaffolds

import (
	"fmt"

	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"github.com/spf13/afero"

	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins"
	kustomizecommonv1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	kustomizecommonv2alpha "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v2-alpha"
	// FGI: overwritting these two lines would allow to inject custom templates
	// "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds/internal/templates"
	// "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3/scaffolds/internal/templates/hack"
	"github.com/fgiloux/kcp-operator-sdk/plugins/golang/v3/scaffolds/internal/templates"
	"github.com/fgiloux/kcp-operator-sdk/plugins/golang/v3/scaffolds/internal/templates/hack"
)

const (
	// ControllerRuntimeVersion is the kubernetes-sigs/controller-runtime version to be used in the project
	// ControllerRuntimeVersion = "v0.13.0"
	// Version currently supported by kcp fork
	ControllerRuntimeVersion = "v0.11.2"
	// ControllerToolsVersion is the kubernetes-sigs/controller-tools version to be used in the project
	ControllerToolsVersion = "v0.10.0"
	KCPVersion             = "0.9.1"
	YQVersion              = "v4.27.2"
	Registry               = "localhost"
	ImageName              = "controller:0.1"
	EnvTestK8s             = "1.25"
)

var _ plugins.Scaffolder = &initScaffolder{}

var kustomizeVersion string

type initScaffolder struct {
	config          config.Config
	boilerplatePath string
	license         string
	owner           string

	// fs is the filesystem that will be used by the scaffolder
	fs machinery.Filesystem
}

// NewInitScaffolder returns a new Scaffolder for project initialization operations
func NewInitScaffolder(config config.Config, license, owner string) plugins.Scaffolder {
	return &initScaffolder{
		config:          config,
		boilerplatePath: hack.DefaultBoilerplatePath,
		license:         license,
		owner:           owner,
	}
}

// InjectFS implements cmdutil.Scaffolder
func (s *initScaffolder) InjectFS(fs machinery.Filesystem) {
	s.fs = fs
}

// Scaffold implements cmdutil.Scaffolder
func (s *initScaffolder) Scaffold() error {
	fmt.Println("Writing scaffold for you to edit...")

	// Initialize the machinery.Scaffold that will write the boilerplate file to disk
	// The boilerplate file needs to be scaffolded as a separate step as it is going to
	// be used by the rest of the files, even those scaffolded in this command call.
	scaffold := machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
	)

	bpFile := &hack.Boilerplate{
		License: s.license,
		Owner:   s.owner,
	}
	bpFile.Path = s.boilerplatePath
	if err := scaffold.Execute(bpFile); err != nil {
		return err
	}

	boilerplate, err := afero.ReadFile(s.fs.FS, s.boilerplatePath)
	if err != nil {
		return err
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold = machinery.NewScaffold(s.fs,
		machinery.WithConfig(s.config),
		machinery.WithBoilerplate(string(boilerplate)),
	)

	// If the KustomizeV2 was used to do the scaffold then
	// we need to ensure that we use its supported Kustomize Version
	// in order to support it
	kustomizeVersion = kustomizecommonv1.KustomizeVersion
	kustomizev2 := kustomizecommonv2alpha.Plugin{}
	pluginKeyForKustomizeV2 := plugin.KeyFor(kustomizev2)

	for _, pluginKey := range s.config.GetPluginChain() {
		if pluginKey == pluginKeyForKustomizeV2 {
			kustomizeVersion = kustomizecommonv2alpha.KustomizeVersion
			break
		}
	}

	return scaffold.Execute(
		&templates.Main{},
		&templates.GoMod{
			ControllerRuntimeVersion: ControllerRuntimeVersion,
		},
		&templates.GitIgnore{},
		&templates.Makefile{
			Registry:                 Registry,
			Image:                    ImageName,
			BoilerplatePath:          s.boilerplatePath,
			ControllerToolsVersion:   ControllerToolsVersion,
			KustomizeVersion:         kustomizeVersion,
			ControllerRuntimeVersion: ControllerRuntimeVersion,
			KCPVersion:               KCPVersion,
			YQVersion:                YQVersion,
			EnvTestK8s:               EnvTestK8s,
		},
		&templates.Dockerfile{},
		&templates.DockerIgnore{},
		&templates.Readme{},
	)
}
