package combustion

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/image"
	"github.com/suse-edge/edge-image-builder/pkg/log"
	"github.com/suse-edge/edge-image-builder/pkg/template"
	"go.uber.org/zap"
)

const (
	k8sComponentName  = "kubernetes"
	rke2InstallScript = "15-rke2-install.sh"
)

var (
	//go:embed templates/15-rke2-installer.sh.tpl
	rke2InstallerScript string
)

func configureKubernetes(ctx *image.Context) ([]string, error) {
	var configureFunc func(*image.Context) (string, error)

	switch ctx.ImageDefinition.Kubernetes.Distribution {
	case "":
		log.AuditComponentSkipped(k8sComponentName)
		return nil, nil
	case image.KubernetesTypeRKE2:
		configureFunc = configureRKE2
	case image.KubernetesTypeK3s:
		panic("implement me")
	default:
		log.AuditComponentFailed(k8sComponentName)
		return nil, fmt.Errorf("unexpected k8s distro: %s", ctx.ImageDefinition.Kubernetes.Distribution)
	}

	script, err := configureFunc(ctx)
	if err != nil {
		log.AuditComponentFailed(k8sComponentName)

		distro := strings.ToUpper(ctx.ImageDefinition.Kubernetes.Distribution)
		return nil, fmt.Errorf("configuring %s: %w", distro, err)
	}

	log.AuditComponentSuccessful(k8sComponentName)
	return []string{script}, nil
}

func installKubernetesScript(ctx *image.Context) error {
	sourcePath := "/" // root level of the container image
	destPath := ctx.CombustionDir

	return ctx.KubernetesScriptInstaller.InstallKubernetesScript(ctx.ImageDefinition.Kubernetes.Distribution, sourcePath, destPath)
}

func configureRKE2(ctx *image.Context) (string, error) {
	if err := installKubernetesScript(ctx); err != nil {
		return "", fmt.Errorf("installing RKE2 script: %w", err)
	}

	configFile, err := copyKubernetesConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("copying RKE2 config: %w", err)
	}

	rke2 := struct {
		image.Kubernetes
		ConfigFile string
	}{
		Kubernetes: ctx.ImageDefinition.Kubernetes,
		ConfigFile: configFile,
	}

	data, err := template.Parse(rke2InstallScript, rke2InstallerScript, &rke2)
	if err != nil {
		return "", fmt.Errorf("parsing RKE2 installer template: %w", err)
	}

	installScript := filepath.Join(ctx.CombustionDir, rke2InstallScript)
	if err = os.WriteFile(installScript, []byte(data), fileio.ExecutablePerms); err != nil {
		return "", fmt.Errorf("writing RKE2 installer script: %w", err)
	}

	return rke2InstallScript, nil
}

func copyKubernetesConfig(ctx *image.Context) (string, error) {
	const (
		k8sConfigDir  = "kubernetes"
		k8sConfigFile = "config.yaml"
	)

	configDir := generateComponentPath(ctx, k8sConfigDir)

	_, err := os.Stat(configDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			zap.S().Info("Kubernetes config file not provided")
			return "", nil
		}

		return "", fmt.Errorf("error checking kubernetes component directory: %w", err)
	}

	configFile := filepath.Join(configDir, k8sConfigFile)

	_, err = os.Stat(configFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("kubernetes component directory exists but does not contain config.yaml")
		}
		return "", fmt.Errorf("error checking kubernetes config file: %w", err)
	}

	destFile := fmt.Sprintf("%s_config.yaml", ctx.ImageDefinition.Kubernetes.Distribution)

	if err = fileio.CopyFile(configFile, filepath.Join(ctx.CombustionDir, destFile), fileio.NonExecutablePerms); err != nil {
		return "", fmt.Errorf("copying kubernetes config file: %w", err)
	}

	return destFile, nil
}
