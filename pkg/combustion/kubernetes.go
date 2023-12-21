package combustion

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/image"
	"github.com/suse-edge/edge-image-builder/pkg/log"
	"github.com/suse-edge/edge-image-builder/pkg/template"
)

const (
	k8sComponentName  = "kubernetes"
	rke2ConfigScript  = "15-rke2-config.sh"
	rke2InstallScript = "15-rke2-install.sh"
	rke2RunScript     = "15-rke2-run.sh"
)

var (
	//go:embed templates/15-rke2-config.sh.tpl
	rke2Config string

	//go:embed templates/15-rke2-installer.sh.tpl
	rke2InstallerScript string

	//go:embed templates/15-rke2-runner.sh.tpl
	rke2RunnerScript string
)

func configureKubernetes(ctx *image.Context) ([]string, error) {
	switch ctx.ImageDefinition.Kubernetes.Distribution {
	case "":
		log.AuditComponentSkipped(k8sComponentName)
		return nil, nil
	case image.KubernetesTypeRKE2:
		if err := configureRKE2(ctx); err != nil {
			return nil, fmt.Errorf("configuring rke2: %w", err)
		}
		log.AuditComponentSuccessful(k8sComponentName)
		return []string{
			rke2ConfigScript,
			rke2InstallScript,
			rke2RunScript,
		}, nil
	case image.KubernetesTypeK3s:
		panic("implement me")
	default:
		log.AuditComponentFailed(k8sComponentName)
		return nil, fmt.Errorf("unexpected k8s distro: %s", ctx.ImageDefinition.Kubernetes.Distribution)
	}
}

func configureRKE2(ctx *image.Context) error {
	data, err := template.Parse(rke2ConfigScript, rke2Config, &ctx.ImageDefinition.Kubernetes)
	if err != nil {
		return fmt.Errorf("parsing RKE2 config template: %w", err)
	}

	configScript := filepath.Join(ctx.CombustionDir, rke2ConfigScript)
	if err = os.WriteFile(configScript, []byte(data), fileio.ExecutablePerms); err != nil {
		return fmt.Errorf("writing RKE2 config: %w", err)
	}

	data, err = template.Parse(rke2InstallScript, rke2InstallerScript, &ctx.ImageDefinition.Kubernetes)
	if err != nil {
		return fmt.Errorf("parsing RKE2 installer template: %w", err)
	}

	installScript := filepath.Join(ctx.CombustionDir, rke2InstallScript)
	if err = os.WriteFile(installScript, []byte(data), fileio.ExecutablePerms); err != nil {
		return fmt.Errorf("writing RKE2 installer script: %w", err)
	}

	data, err = template.Parse(rke2RunScript, rke2RunnerScript, &ctx.ImageDefinition.Kubernetes)
	if err != nil {
		return fmt.Errorf("parsing RKE2 runner template: %w", err)
	}

	runScript := filepath.Join(ctx.CombustionDir, rke2RunScript)
	if err = os.WriteFile(runScript, []byte(data), fileio.ExecutablePerms); err != nil {
		return fmt.Errorf("writing RKE2 runner script: %w", err)
	}

	return nil
}
