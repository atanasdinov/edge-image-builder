package build

import (
	"fmt"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/combustion"
	"github.com/suse-edge/edge-image-builder/pkg/config"
	"github.com/suse-edge/edge-image-builder/pkg/fileio"
)

type Builder struct {
	imageConfig *config.ImageConfig
	context     *Context

	combustionScripts []string
}

func New(imageConfig *config.ImageConfig, context *Context) *Builder {
	return &Builder{
		imageConfig: imageConfig,
		context:     context,
	}
}

func (b *Builder) Build() error {
	err := b.configureMessage()
	if err != nil {
		return fmt.Errorf("configuring the welcome message: %w", err)
	}

	err = b.configureScripts()
	if err != nil {
		return fmt.Errorf("configuring custom scripts: %w", err)
	}

	script, err := combustion.GenerateScript(b.combustionScripts)
	if err != nil {
		return fmt.Errorf("generating combustion script: %w", err)
	}

	if _, err = b.writeCombustionFile("script", script, nil); err != nil {
		return fmt.Errorf("writing combustion script: %w", err)
	}

	err = b.copyRPMs()
	if err != nil {
		return fmt.Errorf("copying RPMs over: %w", err)
	}

	switch b.imageConfig.Image.ImageType {
	case config.ImageTypeISO:
		return b.buildIsoImage()
	case config.ImageTypeRAW:
		return b.buildRawImage()
	default:
		return fmt.Errorf("invalid imageType value specified, must be either \"%s\" or \"%s\"",
			config.ImageTypeISO, config.ImageTypeRAW)
	}
}

func (b *Builder) writeBuildDirFile(filename string, contents string, templateData any) (string, error) {
	destFilename := filepath.Join(b.context.BuildDir, filename)
	return destFilename, fileio.WriteFile(destFilename, contents, templateData)
}

func (b *Builder) writeCombustionFile(filename string, contents string, templateData any) (string, error) {
	destFilename := filepath.Join(b.context.CombustionDir, filename)
	return destFilename, fileio.WriteFile(destFilename, contents, templateData)
}

func (b *Builder) registerCombustionScript(scriptName string) {
	// Keep a running list of all added combustion scripts. When we add the combustion
	// "script" file (the one Combustion itself looks at), we'll concatenate calls to
	// each of these to that script.

	b.combustionScripts = append(b.combustionScripts, scriptName)
}

func (b *Builder) generateOutputImageFilename() string {
	filename := filepath.Join(b.context.ImageConfigDir, b.imageConfig.Image.OutputImageName)
	return filename
}

func (b *Builder) generateBaseImageFilename() string {
	filename := filepath.Join(b.context.ImageConfigDir, "images", b.imageConfig.Image.BaseImage)
	return filename
}
