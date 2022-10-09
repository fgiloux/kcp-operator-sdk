package golang

import "github.com/fgiloux/kcp-operator-sdk/plugins"

// DefaultNameQualifier is the suffix appended to all kubebuilder plugin names for Golang operators.
const DefaultNameQualifier = "go." + plugins.DefaultNameQualifier
