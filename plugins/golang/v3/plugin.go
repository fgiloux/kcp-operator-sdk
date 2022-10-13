package v3

import (
	"github.com/fgiloux/kcp-operator-sdk/plugins/golang"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	cfgv3 "sigs.k8s.io/kubebuilder/v3/pkg/config/v3"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"
)

const pluginName = "base." + golang.DefaultNameQualifier

var (
	pluginVersion            = plugin.Version{Number: 3}
	supportedProjectVersions = []config.Version{cfgv3.Version}
)

// FGI: starting simple
// var _ plugin.Full = Plugin{}

// Plugin implements the plugin.Full interface
type Plugin struct {
	initSubcommand
	createAPISubcommand
	// createWebhookSubcommand
	// editSubcommand
}

// Name returns the name of the plugin
func (Plugin) Name() string { return pluginName }

// Version returns the version of the plugin
func (Plugin) Version() plugin.Version { return pluginVersion }

// SupportedProjectVersions returns an array with all project versions supported by the plugin
func (Plugin) SupportedProjectVersions() []config.Version { return supportedProjectVersions }

// GetInitSubcommand will return the subcommand which is responsible for initializing and common scaffolding
func (p Plugin) GetInitSubcommand() plugin.InitSubcommand { return &p.initSubcommand }

// GetCreateAPISubcommand will return the subcommand which is responsible for scaffolding apis
func (p Plugin) GetCreateAPISubcommand() plugin.CreateAPISubcommand { return &p.createAPISubcommand }

// GetCreateWebhookSubcommand will return the subcommand which is responsible for scaffolding webhooks
/*func (p Plugin) GetCreateWebhookSubcommand() plugin.CreateWebhookSubcommand {
	return &p.createWebhookSubcommand
}*/

// GetEditSubcommand will return the subcommand which is responsible for editing the scaffold of the project
// func (p Plugin) GetEditSubcommand() plugin.EditSubcommand { return &p.editSubcommand }
