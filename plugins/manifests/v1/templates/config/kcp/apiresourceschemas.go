package kcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &APIResourceschemas{}

// APIResourceschemas scaffolds an apiresourceschemas.yaml for the manifests overlay folder.
type APIResourceschemas struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *APIResourceschemas) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "kcp", "apiresourceschemas.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = apirsTemplate

	return nil
}

const apirsTemplate = `# APIResourceschemas for the controller custom resources
`
