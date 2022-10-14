package defaultkcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Kustomization{}

// Kustomization scaffolds a kustomization.yaml for the manifests overlay folder.
type Kustomization struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *Kustomization) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "default-kcp", "kustomization.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = kustomizationTemplate

	return nil
}

const kustomizationTemplate = `# These resources are the kcp specific manifests
# Adds namespace to all resources.
namespace: {{ .ProjectName }}-system

# Value of this field is prepended to the
# names of all resources, e.g. a deployment named
# "wordpress" becomes "alices-wordpress".
# Note that it should also match with the prefix (text before '-') of the namespace
# field above.
namePrefix: {{ .ProjectName }}-

# Labels to add to all resources and selectors.
#commonLabels:
#  someName: someValue

bases:
- ../kcp
- ../rbac
- ../manager

patchesStrategicMerge:
- manager_patch.yaml

configurations:
- kustomizeconfig.yaml

# Adjust to prefix
vars:
- name: API_EXPORT_NAME
  objref:
    apiVersion: apis.kcp.dev/v1alpha1
    kind: APIExport
    name: {{ .ProjectName }}.{{ .Domain }}
  fieldref:
    fieldPath: metadata.name
`
