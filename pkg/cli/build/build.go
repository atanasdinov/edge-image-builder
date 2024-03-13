package build

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/build"
	"github.com/suse-edge/edge-image-builder/pkg/cache"
	"github.com/suse-edge/edge-image-builder/pkg/cli/cmd"
	"github.com/suse-edge/edge-image-builder/pkg/combustion"
	"github.com/suse-edge/edge-image-builder/pkg/helm"
	"github.com/suse-edge/edge-image-builder/pkg/image"
	"github.com/suse-edge/edge-image-builder/pkg/kubernetes"
	"github.com/suse-edge/edge-image-builder/pkg/log"
	"github.com/suse-edge/edge-image-builder/pkg/network"
	"github.com/suse-edge/edge-image-builder/pkg/podman"
	"github.com/suse-edge/edge-image-builder/pkg/rpm"
	"github.com/suse-edge/edge-image-builder/pkg/rpm/resolver"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	buildLogFilename = "eib-build.log"
	checkLogMessage  = "Please check the eib-build.log file under the build directory for more information."
)

func Run(_ *cli.Context) error {
	args := &cmd.BuildArgs

	rootBuildDir := args.RootBuildDir
	if rootBuildDir == "" {
		const defaultBuildDir = "_build"

		rootBuildDir = filepath.Join(args.ConfigDir, defaultBuildDir)
		if err := os.MkdirAll(rootBuildDir, os.ModePerm); err != nil {
			log.Auditf("The root build directory could not be set up under the configuration directory '%s'.", args.ConfigDir)
			return err
		}
	}

	buildDir, combustionDir, err := build.SetupBuildDirectory(rootBuildDir)
	if err != nil {
		log.Audit("The build directory could not be set up.")
		return err
	}

	// This needs to occur as early as possible so that the subsequent calls can use the log
	log.ConfigureGlobalLogger(filepath.Join(buildDir, buildLogFilename))

	configDirExists := imageConfigDirExists(args.ConfigDir)
	if !configDirExists {
		os.Exit(1)
	}

	imageDefinition := parseImageDefinition(args.ConfigDir, args.DefinitionFile)
	if imageDefinition == nil {
		os.Exit(1)
	}

	ctx := buildContext(buildDir, combustionDir, args.ConfigDir, imageDefinition)

	isDefinitionValid := isImageDefinitionValid(ctx)
	if !isDefinitionValid {
		os.Exit(1)
	}

	if err = appendKubernetesSELinuxRPMs(ctx); err != nil {
		log.Auditf("Configuring Kubernetes failed. %s", checkLogMessage)
		zap.S().Fatalf("Failed to configure Kubernetes SELinux policy: %s", err)
	}

	appendElementalRPMs(ctx)

	appendHelm(ctx)

	if !bootstrapDependencyServices(ctx, rootBuildDir) {
		os.Exit(1)
	}

	defer func() {
		if r := recover(); r != nil {
			log.AuditInfo("Build failed unexpectedly, check the logs under the build directory for more information.")
			zap.S().Fatalf("Unexpected error occurred: %s", r)
		}
	}()

	builder := build.NewBuilder(ctx)
	if err = builder.Build(); err != nil {
		zap.S().Fatalf("An error occurred building the image: %s", err)
	}

	return nil
}

// Returns whether the image configuration directory can be read, displaying
// the appropriate messages to the user. Returns 'true' if the directory exists and execution can proceed,
// 'false' otherwise.
func imageConfigDirExists(configDir string) bool {
	_, err := os.Stat(configDir)
	if err == nil {
		return true
	}

	if errors.Is(err, fs.ErrNotExist) {
		log.AuditInfof("The specified image configuration directory '%s' could not be found.", configDir)
	} else {
		log.AuditInfof("Unable to check the filesystem for the image configuration directory '%s'. %s",
			configDir, checkLogMessage)

		zap.S().Errorf("Reading config dir failed: %v", err)
	}

	return false
}

// Attempts to parse the specified image definition file, displaying the appropriate messages to the user.
// Returns a populated `image.Context` struct if successful, `nil` if the definition could not be parsed.
func parseImageDefinition(configDir, definitionFile string) *image.Definition {
	definitionFilePath := filepath.Join(configDir, definitionFile)

	configData, err := os.ReadFile(definitionFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			log.AuditInfof("The specified definition file '%s' could not be found.", definitionFilePath)
		} else {
			log.AuditInfof("The specified definition file '%s' could not be read. %s", definitionFilePath, checkLogMessage)
			zap.S().Error(err)
		}
		return nil
	}

	imageDefinition, err := image.ParseDefinition(configData)
	if err != nil {
		log.AuditInfof("The image definition file '%s' could not be parsed. %s", definitionFilePath, checkLogMessage)
		zap.S().Error(err)
		return nil
	}

	return imageDefinition
}

// Assembles the image build context with user-provided values and implementation defaults.
func buildContext(buildDir, combustionDir, configDir string, imageDefinition *image.Definition) *image.Context {
	ctx := &image.Context{
		ImageConfigDir:               configDir,
		BuildDir:                     buildDir,
		CombustionDir:                combustionDir,
		ImageDefinition:              imageDefinition,
		NetworkConfigGenerator:       network.ConfigGenerator{},
		NetworkConfiguratorInstaller: network.ConfiguratorInstaller{},
	}
	return ctx
}

func appendKubernetesSELinuxRPMs(ctx *image.Context) error {
	if ctx.ImageDefinition.Kubernetes.Version == "" {
		return nil
	}

	configPath := combustion.KubernetesConfigPath(ctx)
	config, err := kubernetes.ParseKubernetesConfig(configPath)
	if err != nil {
		return fmt.Errorf("parsing kubernetes server config: %w", err)
	}

	selinuxEnabled, _ := config["selinux"].(bool)
	if !selinuxEnabled {
		return nil
	}

	log.AuditInfo("SELinux is enabled in the Kubernetes configuration. " +
		"The necessary RPM packages will be downloaded.")

	selinuxPackage, err := kubernetes.SELinuxPackage(ctx.ImageDefinition.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("identifying selinux package: %w", err)
	}

	repository, err := kubernetes.SELinuxRepository(ctx.ImageDefinition.Kubernetes.Version)
	if err != nil {
		return fmt.Errorf("identifying selinux repository: %w", err)
	}

	appendRPMs(ctx, repository, selinuxPackage)

	gpgKeysDir := combustion.GPGKeysPath(ctx)
	if err = os.MkdirAll(gpgKeysDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory '%s': %w", gpgKeysDir, err)
	}

	if err = kubernetes.DownloadSELinuxRPMsSigningKey(gpgKeysDir); err != nil {
		return fmt.Errorf("downloading signing key: %w", err)
	}

	return nil
}

func appendElementalRPMs(ctx *image.Context) {
	elementalDir := combustion.ElementalPath(ctx)
	if _, err := os.Stat(elementalDir); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			zap.S().Warnf("Looking for '%s' dir failed unexpectedly: %s", elementalDir, err)
		}

		return
	}

	log.AuditInfo("Elemental registration is configured. The necessary RPM packages will be downloaded.")

	appendRPMs(ctx, image.AddRepo{URL: combustion.ElementalPackageRepository}, combustion.ElementalPackages...)
}

func appendRPMs(ctx *image.Context, repository image.AddRepo, packages ...string) {
	repositories := ctx.ImageDefinition.OperatingSystem.Packages.AdditionalRepos
	repositories = append(repositories, repository)

	packageList := ctx.ImageDefinition.OperatingSystem.Packages.PKGList
	packageList = append(packageList, packages...)

	ctx.ImageDefinition.OperatingSystem.Packages.PKGList = packageList
	ctx.ImageDefinition.OperatingSystem.Packages.AdditionalRepos = repositories
}

func appendHelm(ctx *image.Context) {
	componentCharts, componentRepos := combustion.ComponentHelmCharts(ctx)

	ctx.ImageDefinition.Kubernetes.Helm.Charts = append(ctx.ImageDefinition.Kubernetes.Helm.Charts, componentCharts...)
	ctx.ImageDefinition.Kubernetes.Helm.Repositories = append(ctx.ImageDefinition.Kubernetes.Helm.Repositories, componentRepos...)
}

// If the image definition requires it, starts the necessary services, displaying appropriate messages
// to users in the event of an error. Returns 'true' if execution should proceed given that all dependencies
// are satisfied; 'false' otherwise.
func bootstrapDependencyServices(ctx *image.Context, rootDir string) bool {
	if !combustion.SkipRPMComponent(ctx) {
		p, err := podman.New(ctx.BuildDir)
		if err != nil {
			log.AuditInfof("The services for RPM dependency resolution failed to start. %s", checkLogMessage)
			zap.S().Error(err)
			return false
		}

		imgPath := filepath.Join(ctx.ImageConfigDir, "base-images", ctx.ImageDefinition.Image.BaseImage)
		imgType := ctx.ImageDefinition.Image.ImageType
		baseBuilder := resolver.NewTarballBuilder(ctx.BuildDir, imgPath, imgType, p)

		rpmResolver := resolver.New(ctx.BuildDir, p, baseBuilder, "")
		ctx.RPMResolver = rpmResolver
		ctx.RPMRepoCreator = rpm.NewRepoCreator(ctx.BuildDir)
	}

	if combustion.IsEmbeddedArtifactRegistryConfigured(ctx) {
		ctx.HelmClient = helm.New(ctx.BuildDir)
	}

	if ctx.ImageDefinition.Kubernetes.Version != "" {
		c, err := cache.New(rootDir)
		if err != nil {
			log.AuditInfof("Failed to initialise file caching. %s", checkLogMessage)
			zap.S().Error(err)
			return false
		}

		ctx.KubernetesScriptDownloader = kubernetes.ScriptDownloader{}
		ctx.KubernetesArtefactDownloader = kubernetes.ArtefactDownloader{
			Cache: c,
		}
	}

	return true
}
