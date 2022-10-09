package templates

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &GitIgnore{}

// GitIgnore scaffolds a file that defines which files should be ignored by git
type GitIgnore struct {
	machinery.TemplateMixin
}

// SetTemplateDefaults implements file.Template
func (f *GitIgnore) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = ".gitignore"
	}

	f.TemplateBody = gitignoreTemplate

	return nil
}

const gitignoreTemplate = `
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin
testbin/*
Dockerfile.cross

# Test binary, build with ` + "`go test -c`" + `
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories
vendor/

# editor and IDE paraphernalia
.idea
*.swp
*.swo
*~
`
