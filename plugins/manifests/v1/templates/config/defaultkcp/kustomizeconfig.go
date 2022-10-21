package defaultkcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Kustomization{}

// KustomizeConfig scaffolds a kustomizeconfig.yaml for the manifests overlay folder.
type KustomizeConfig struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *KustomizeConfig) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "default-kcp", "kustomizeconfig.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = kustomizeConfigTemplate

	return nil
}

const kustomizeConfigTemplate = `nameReference:
- kind: APIResourceSchema
  fieldSpecs:
  - kind: APIExport
    path: spec/latestResourceSchemas
- kind: ConfigMap
  fieldSpecs:
  - kind: Deployment
    path: spec/volumes/configMap/name

`
