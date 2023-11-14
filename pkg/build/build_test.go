package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteCombustionFile(t *testing.T) {
	// Setup
	context, err := NewContext("", "", true)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, CleanUpBuildDir(context))
	}()

	builder := Builder{
		context: context,
	}

	testData := "Edge Image Builder"
	testFilename := "combustion-file.sh"

	// Test
	writtenFilename, err := builder.writeCombustionFile(testFilename, testData, nil)

	// Verify
	require.NoError(t, err)

	expectedFilename := filepath.Join(context.CombustionDir, testFilename)
	foundData, err := os.ReadFile(expectedFilename)
	require.NoError(t, err)
	assert.Equal(t, expectedFilename, writtenFilename)
	assert.Equal(t, testData, string(foundData))
}

func TestWriteBuildDirFile(t *testing.T) {
	// Setup
	context, err := NewContext("", "", true)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, CleanUpBuildDir(context))
	}()

	builder := Builder{
		context: context,
	}

	testData := "Edge Image Builder"
	testFilename := "build-dir-file.sh"

	// Test
	writtenFilename, err := builder.writeBuildDirFile(testFilename, testData, nil)

	// Verify
	require.NoError(t, err)

	expectedFilename := filepath.Join(context.BuildDir, testFilename)
	require.Equal(t, expectedFilename, writtenFilename)
	foundData, err := os.ReadFile(expectedFilename)
	require.NoError(t, err)
	assert.Equal(t, testData, string(foundData))
}
