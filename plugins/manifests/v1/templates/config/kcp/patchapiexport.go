package kcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &PatchAPIExport{}

// PatchAPIExport scaffolds a kustomizeconfig.yaml for the manifests overlay folder.
type PatchAPIExport struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin
	machinery.ResourceMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *PatchAPIExport) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "kcp", "patch_apiexport.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = patchAPIExportTemplate

	return nil
}

const patchAPIExportTemplate = `# Set the reference to the latest APIRresourceSchema
---
apiVersion: apis.kcp.dev/v1alpha1
kind: APIExport
metadata:
  name: {{ .ProjectName }}.{{ .Domain }}
spec:
  latestResourceSchemas:
     - PREFIX.{{ .Resource.Plural }}.{{ .Domain }}

`
