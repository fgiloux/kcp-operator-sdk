package e2e

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Audit{}

// Audit scaffolds an audit policy for the e2e tests.
type Audit struct {
	machinery.TemplateMixin
}

// SetTemplateDefaults implements machinery.Template
func (f *Audit) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("test", "e2e", "audit-policy.yaml")
	}

	// We cannot overwrite the file after it has been created because
	// user changes may get lost, i.e to work with Kustomize 4.x
	// the target /spec/template/spec/containers/1/volumeMounts/0
	// needs to be replaced with /spec/template/spec/containers/0/volumeMounts/0
	f.IfExistsAction = machinery.SkipFile

	f.TemplateBody = auditTemplate

	return nil
}

const auditTemplate = `---
apiVersion: audit.k8s.io/v1
kind: Policy
omitStages:
  - RequestReceived
omitManagedFields: true
rules:
  - level: None
    nonResourceURLs:
      - "/api*"
      - "/version"

  - level: Metadata
    resources:
      - group: ""
        resources: ["secrets", "configmaps"]
      - group: "authorization.k8s.io"
        resources: ["subjectaccessreviews"]

  - level: Metadata
    verbs: ["list", "watch"]

  - level: Metadata
    verbs: ["get", "delete"]
    omitStages:
      - ResponseStarted

  - level: RequestResponse
    verbs: ["create", "update", "patch"]
    omitStages:
      - ResponseStarted

`
