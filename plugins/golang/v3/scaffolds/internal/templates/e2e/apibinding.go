package e2e

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &APIBinding{}

// APIBinding scaffolds an apibinding.yaml for the e2e tests.
// This is needed as long as a controller can not start
// watching resources of an APIExport for which no APIBinding has been created.
type APIBinding struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *APIBinding) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("test", "e2e", "apibinding.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = apibindingTemplate

	return nil
}

const apibindingTemplate = `---
apiVersion: apis.kcp.dev/v1alpha1
kind: APIBinding
metadata:
  name: {{ .ProjectName }}-{{ .ProjectName }}.{{ .Domain }}
spec:
  reference:
    workspace:
      path: WORKSPACE
      exportName: {{ .ProjectName }}-{{ .ProjectName }}.{{ .Domain }}
  permissionClaims:
  # TODO (user)

`
