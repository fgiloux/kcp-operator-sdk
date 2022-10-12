package kcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &APIExport{}

// APIExport scaffolds an apiexport.yaml for the manifests overlay folder.
type APIExport struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *APIExport) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "kcp", "apiexport.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = apiexportTemplate

	return nil
}

const apiexportTemplate = `# Controller APIExport
apiVersion: apis.kcp.dev/v1alpha1
kind: APIExport
metadata:
  name: {{ .ProjectName }}.{{ .Domain }}
spec:
`
