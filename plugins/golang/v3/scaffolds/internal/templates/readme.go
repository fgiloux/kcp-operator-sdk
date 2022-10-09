package templates

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Readme{}

// Readme scaffolds a README.md file
type Readme struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.ProjectNameMixin

	License string
}

// SetTemplateDefaults implements file.Template
func (f *Readme) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "README.md"
	}

	f.License = strings.Replace(
		strings.Replace(f.Boilerplate, "/*", "", 1),
		"*/", "", 1)

	f.TemplateBody = fmt.Sprintf(readmeFileTemplate,
		codeFence("make docker-build docker-push REGISTRY=<some-registry> IMG={{ .ProjectName }}:tag"),
		codeFence("make deploy REGISTRY=<some-registry> IMG={{ .ProjectName }}:tag"),
		codeFence("make uninstall"),
		codeFence("make undeploy"),
		codeFence("make install"),
		codeFence("make run"),
		codeFence("make manifests apiresourceschemas"))

	return nil
}

//nolint:lll
const readmeFileTemplate = `# {{ .ProjectName }}

// TODO(user): A simple overview of the project and its purpose.

## Description

// TODO(user): An in-depth paragraph providing more details about the project and its use.

## Getting Started

You’ll need a Kubernetes and optionally a kcp cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster ` + "`kubectl cluster-info`" + ` shows).

### Running on Kubernetes or kcp

1. Build and push your image to the location specified by ` + "`REGISTRY` and `IMG`" + `:
	
%s
	
2. Deploy the controller to the cluster with the image specified by ` + "`REGISTRY` and `IMG`" + `:

%s

### Uninstall resources

To delete the resources from the cluster:

%s

### Undeploy controller

Undeploy the controller from the cluster:

%s

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project.

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached. 

### Test It Out

1. Install the required resources into the cluster:

%s

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

%s

**NOTE:** You can also run this in one step by running: ` + "`make install run`" + `

### Modifying the API definitions

If you are editing the API definitions, regenerate the manifests using:

%s

**NOTE:** Run ` + "`make --help`" + ` for more information on all potential ` + "`make`" + ` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

{{ .License }}
`

func codeFence(code string) string {
	return "```sh" + "\n" + code + "\n" + "```"
}
