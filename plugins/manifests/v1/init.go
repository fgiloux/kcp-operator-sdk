package v1

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/plugin"

	"github.com/fgiloux/kcp-operator-sdk/plugins/manifests/v1/templates/config/defaultkcp"
	kcptemplates "github.com/fgiloux/kcp-operator-sdk/plugins/manifests/v1/templates/config/kcp"
)

const filePath = "Makefile"

var _ plugin.InitSubcommand = &initSubcommand{}

type initSubcommand struct {
	config config.Config
}

func (s *initSubcommand) InjectConfig(c config.Config) error {
	s.config = c

	return nil
}

func (s *initSubcommand) Scaffold(fs machinery.Filesystem) error {

	makefileBytes, err := afero.ReadFile(fs.FS, filePath)
	if err != nil {
		return err
	}

	// Prepend bundle variables.
	projectName := s.config.GetProjectName()
	if projectName == "" {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}
		projectName = strings.ToLower(filepath.Base(dir))
	}
	/*makefileBytes = append([]byte(fmt.Sprintf(makefileBundleVarFragment, s.config.GetDomain(), projectName)), makefileBytes...)

	// Append bundle recipes.
	makefileBytes = append(makefileBytes, []byte(makefileBundleFragmentGo)...)
	makefileBytes = append(makefileBytes, []byte(makefileBundleBuildPushFragment)...)

	// Append catalog recipes.
	makefileBytes = append(makefileBytes, []byte(fmt.Sprintf(makefileOPMFragmentGo, "to be replaced"))...)
	makefileBytes = append(makefileBytes, []byte(makefileCatalogBuildFragment)...)
	*/

	var mode os.FileMode = 0644
	if info, err := fs.FS.Stat(filePath); err == nil {
		mode = info.Mode()
	}
	if err := afero.WriteFile(fs.FS, filePath, makefileBytes, mode); err != nil {
		return fmt.Errorf("error updating Makefile: %w", err)
	}

	// Initialize the machinery.Scaffold that will write the files to disk
	scaffold := machinery.NewScaffold(fs,
		// NOTE: kubebuilder's default permissions are only for root users
		machinery.WithDirectoryPermissions(0755),
		machinery.WithFilePermissions(0644),
		machinery.WithConfig(s.config),
	)

	if err := scaffold.Execute(
		&kcptemplates.Kustomization{},
		&kcptemplates.Clusterrolebinding{},
		&kcptemplates.Clusterrole{},
		&defaultkcp.Kustomization{},
		&defaultkcp.KustomizeConfig{},
		&defaultkcp.ManagerPatch{},
	); err != nil {
		return fmt.Errorf("error scaffolding manifests: %w", err)
	}

	if err := s.config.EncodePluginConfig(pluginKey, Config{}); err != nil && !errors.As(err, &config.UnsupportedFieldError{}) {
		return err
	}

	return nil
}
