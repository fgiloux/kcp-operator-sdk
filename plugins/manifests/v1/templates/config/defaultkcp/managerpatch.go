package defaultkcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &ManagerPatch{}

// ManagerPatch scaffolds a manager_patch.yaml for the manifests overlay folder.
type ManagerPatch struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
	machinery.DomainMixin
	machinery.ComponentConfigMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *ManagerPatch) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "default-kcp", "manager_patch.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = mgrPatchTemplate

	return nil
}

const mgrPatchTemplate = `# Pass the name of the APIExport to the controller
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - name: manager
        args:
        - "--api-export-name=$(API_EXPORT_NAME)"
{{- if not .ComponentConfig }}
        - --leader-elect
{{- end }}

`
