package kcp

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Clusterrole{}

// Clusterrole scaffolds a clusterrole.yaml for the manifests overlay folder.
type Clusterrole struct {
	machinery.TemplateMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *Clusterrole) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "kcp", "clusterrole.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = crTemplate

	return nil
}

const crTemplate = `# This contains the rights required by the controller
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: kcp-manager-role
rules:
- apiGroups:
  - apis.kcp.dev
  resources:
  - apiexports
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apis.kcp.dev
  resources:
  - apiexports/content
  verbs:
  - '*'

`
