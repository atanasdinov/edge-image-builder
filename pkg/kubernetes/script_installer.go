package kubernetes

import (
	"fmt"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
)

type ScriptInstaller struct{}

func (i ScriptInstaller) InstallKubernetesScript(distro, sourcePath, destinationPath string) error {
	if distro != "k3s" && distro != "rke2" {
		return fmt.Errorf("unsupported distro: %s", distro)
	}

	installerScript := fmt.Sprintf("%s_installer.sh", distro)

	installerSource := filepath.Join(sourcePath, installerScript)
	installerDestination := filepath.Join(destinationPath, installerScript)

	if err := fileio.CopyFile(installerSource, installerDestination, fileio.ExecutablePerms); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}

	return nil
}
