package kcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Clusterrolebinding{}

// Clusterrolebinding scaffolds a clusterrolebinding.yaml for the manifests overlay folder.
type Clusterrolebinding struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *Clusterrolebinding) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "kcp", "clusterrolebinding.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = crbTemplate

	return nil
}

const crbTemplate = `# This contains the clusterrolebinding for the controller
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kcp-manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kcp-manager-role
subjects:
- kind: ServiceAccount
  name: controller-manager
  namespace: system

`
