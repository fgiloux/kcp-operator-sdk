package api

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Group{}

// Group scaffolds the file that defines the registration methods for a certain group and version
type Group struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin
}

// SetTemplateDefaults implements file.Template
func (f *Group) SetTemplateDefaults() error {
	if f.Path == "" {
		if f.MultiGroup {
			if f.Resource.Group != "" {
				f.Path = filepath.Join("apis", "%[group]", "%[version]", "groupversion_info.go")
			} else {
				f.Path = filepath.Join("apis", "%[version]", "groupversion_info.go")
			}
		} else {
			f.Path = filepath.Join("api", "%[version]", "groupversion_info.go")
		}
	}
	f.Path = f.Resource.Replacer().Replace(f.Path)

	f.TemplateBody = groupTemplate

	return nil
}

//nolint:lll
const groupTemplate = `{{ .Boilerplate }}

// Package {{ .Resource.Version }} contains API Schema definitions for the {{ .Resource.Group }} {{ .Resource.Version }} API group
//+kubebuilder:object:generate=true
//+groupName={{ .Resource.QualifiedGroup }}
package {{ .Resource.Version }}

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "{{ .Resource.QualifiedGroup }}", Version: "{{ .Resource.Version }}"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
`