package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/fileio"
)

func TestScriptInstaller_InstallKubernetesScript(t *testing.T) {
	k3sScriptContents := []byte("k3s")
	rke2ScriptContents := []byte("rke2")

	srcDir, err := os.MkdirTemp("", "eib-kubernetes-installer-source-")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(srcDir))
	}()

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "k3s_installer.sh"), k3sScriptContents, fileio.NonExecutablePerms))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "rke2_installer.sh"), rke2ScriptContents, fileio.NonExecutablePerms))

	destDir, err := os.MkdirTemp("", "eib-kubernetes-installer-dest-")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(destDir))
	}()

	tests := []struct {
		name             string
		distro           string
		sourcePath       string
		expectedContents []byte
		expectedError    string
	}{
		{
			name:          "Invalid kubernetes distribution",
			distro:        "rke",
			expectedError: "unsupported distro: rke",
		},
		{
			name:          "Failure to copy non-existing script",
			distro:        "k3s",
			sourcePath:    "",
			expectedError: "copying file: opening source file: open k3s_installer.sh: no such file or directory",
		},
		{
			name:             "Successfully installed k3s script",
			distro:           "k3s",
			sourcePath:       srcDir,
			expectedContents: k3sScriptContents,
		},
		{
			name:             "Successfully installed rke2 script",
			distro:           "rke2",
			sourcePath:       srcDir,
			expectedContents: rke2ScriptContents,
		},
	}

	var installer ScriptInstaller

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err = installer.InstallKubernetesScript(test.distro, test.sourcePath, destDir)

			if test.expectedError != "" {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)

			destPath := filepath.Join(destDir, fmt.Sprintf("%s_installer.sh", test.distro))

			contents, err := os.ReadFile(destPath)
			require.NoError(t, err)
			assert.Equal(t, test.expectedContents, contents)

			info, err := os.Stat(destPath)
			require.NoError(t, err)
			assert.Equal(t, fileio.ExecutablePerms, info.Mode())
		})
	}
}
