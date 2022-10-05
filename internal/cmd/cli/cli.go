package cli

import (
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kubebuilder/v3/pkg/cli"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
	kustomizev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/common/kustomize/v1"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang"
	declarativev1 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/declarative/v1"
	golangv3 "sigs.k8s.io/kubebuilder/v3/pkg/plugins/golang/v3"

	"github.com/fgiloux/kcp-operator-sdk/internal/version"
)

var (
	// This would be where additional specific
	// to this project binary may get added
	// example: myExampleCommand.NewCmd(),
	commands      = []*cobra.Command{}
	alphaCommands = []*cobra.Command{}
)

func Run() error {
	c := GetPluginsCLI()
	return c.Run()
}

// GetPluginsCLI returns the plugins based CLI configured to be used in the new CLI binary
func GetPluginsCLI() *cli.CLI {
	// Bundle plugin which built the golang projects scaffold by Kubebuilder go/v3
	gov3Bundle, _ := plugin.NewBundle(golang.DefaultNameQualifier, plugin.Version{Number: 3},
		kustomizev1.Plugin{},
		golangv3.Plugin{},
	)

	c, err := cli.New(
		// Add the name of your CLI binary
		cli.WithCommandName("kcp-operator-sdk"),

		// Add the version of your CLI binary
		cli.WithVersion(versionString()),

		// Register the plugins options which can be used to do the scaffolds via your CLI tool. See that we are using as example here the plugins which are implemented and provided by Kubebuilder
		cli.WithPlugins(
			gov3Bundle,
			&declarativev1.Plugin{},
		),

		// Defines what will be the default plugin used by your binary. It means that will be the plugin used if no info be provided such as when the user runs `kubebuilder init`
		cli.WithDefaultPlugins(cfgv3.Version, gov3Bundle),

		// Define the default project configuration version which will be used by the CLI when none is informed by --project-version flag.
		cli.WithDefaultProjectVersion(cfgv3.Version),

		// Adds your own commands to the CLI
		cli.WithExtraCommands(commands...),

		// Add your own alpha commands to the CLI
		cli.WithExtraAlphaCommands(alphaCommands...),

		// Adds the completion option for your CLI
		cli.WithCompletion(),
	)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

// versionString returns the CLI version
func versionString() string {
	return fmt.Sprintf("kcp-operator-sdk version: %q, commit: %q, kubernetes version: %q, go version: %q, GOOS: %q, GOARCH: %q",
		version.GitVersion,
		version.GitCommit,
		version.KubernetesVersion,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH)
}
