package combustion

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/image"
)

type mockKubernetesScriptInstaller struct {
	installKubernetesScript func(distro, sourcePath, destPath string) error
}

func (m mockKubernetesScriptInstaller) InstallKubernetesScript(distro, sourcePath, destPath string) error {
	if m.installKubernetesScript != nil {
		return m.installKubernetesScript(distro, sourcePath, destPath)
	}

	panic("not implemented")
}

func TestConfigureKubernetes_Skipped(t *testing.T) {
	ctx := &image.Context{
		ImageDefinition: &image.Definition{},
	}

	scripts, err := configureKubernetes(ctx)
	require.NoError(t, err)
	assert.Nil(t, scripts)
}

func TestConfigureKubernetes_UnsupportedDistro(t *testing.T) {
	ctx := &image.Context{
		ImageDefinition: &image.Definition{
			Kubernetes: image.Kubernetes{
				Distribution: "rke",
			},
		},
	}

	scripts, err := configureKubernetes(ctx)
	require.Error(t, err)
	assert.EqualError(t, err, "unexpected k8s distro: rke")
	assert.Nil(t, scripts)
}

func TestConfigureRKE2(t *testing.T) {
	tests := []struct {
		name            string
		scriptInstaller mockKubernetesScriptInstaller
		expectedErr     string
	}{
		{
			name: "Installing RKE2 script fails",
			scriptInstaller: mockKubernetesScriptInstaller{
				installKubernetesScript: func(distro, sourcePath, destPath string) error {
					return fmt.Errorf("some error")
				},
			},
			expectedErr: "installing RKE2 script: some error",
		},
		{
			name: "Successful configuration",
			scriptInstaller: mockKubernetesScriptInstaller{
				installKubernetesScript: func(distro, sourcePath, destPath string) error {
					return nil
				},
			},
			expectedErr: "",
		},
	}

	ctx, teardown := setupContext(t)
	defer teardown()

	ctx.ImageDefinition.Kubernetes.Distribution = "rke2"

	for _, test := range tests {
		ctx.KubernetesScriptInstaller = test.scriptInstaller

		t.Run(test.name, func(t *testing.T) {
			err := configureRKE2(ctx)

			if test.expectedErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr)
				return
			}

			// TODO: Verify output scripts' contents
			assert.NoError(t, err)
		})
	}
}
