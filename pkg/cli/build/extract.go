package build

import (
	"fmt"
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
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	extractLogFilename     = "eib-extract.log"
	checkExtractLogMessage = "Please check the eib-extract.log file under the build directory for more information."
)

func Extract(_ *cli.Context) error {
	args := &cmd.BuildArgs

	rootBuildDir := args.RootBuildDir
	if rootBuildDir == "" {
		const defaultBuildDir = "_extract"

		rootBuildDir = filepath.Join(args.ConfigDir, defaultBuildDir)
		if err := os.MkdirAll(rootBuildDir, os.ModePerm); err != nil {
			log.Auditf("The root extract directory could not be set up under the configuration directory '%s'.", args.ConfigDir)
			return err
		}
	}

	buildDir, err := build.SetupBuildDirectory(rootBuildDir)
	if err != nil {
		log.Audit("The extract directory could not be set up.")
		return err
	}

	// This needs to occur as early as possible so that the subsequent calls can use the log
	log.ConfigureGlobalLogger(filepath.Join(buildDir, extractLogFilename))

	if cmdErr := imageConfigDirExists(args.ConfigDir); cmdErr != nil {
		cmd.LogError(cmdErr, checkExtractLogMessage)
		os.Exit(1)
	}

	imageDefinition, cmdErr := parseImageDefinition(args.ConfigDir, args.DefinitionFile)
	if cmdErr != nil {
		cmd.LogError(cmdErr, checkExtractLogMessage)
		os.Exit(1)
	}

	artefactsDir, err := build.SetupExtractArtifactsDirectory(buildDir)
	if err != nil {
		log.Auditf("Setting up the extract artifacts directory failed. %s", checkExtractLogMessage)
		zap.S().Fatalf("Failed to create extract artifacts directories: %s", err)
	}

	ctx := buildContext(buildDir, "", artefactsDir, args.ConfigDir, imageDefinition)
	if cmdErr = validateImageDefinition(ctx); cmdErr != nil {
		cmd.LogError(cmdErr, checkExtractLogMessage)
		os.Exit(1)
	}

	appendHelm(ctx)

	if cmdErr = bootstrapExtractDependencyServices(ctx, rootBuildDir); cmdErr != nil {
		cmd.LogError(cmdErr, checkExtractLogMessage)
		os.Exit(1)
	}

	defer func() {
		if r := recover(); r != nil {
			log.AuditInfo("Extract failed unexpectedly, check the logs under the extract directory for more information.")
			zap.S().Fatalf("Unexpected error occurred: %s", r)
		}
	}()

	builder := build.NewBuilder(ctx)
	if err = builder.Extract(); err != nil {
		zap.S().Fatalf("An error occurred extracting the image: %s", err)
	}

	return nil
}

// If the image definition requires it, starts the necessary services, returning an error in the event of failure.
func bootstrapExtractDependencyServices(ctx *image.Context, rootDir string) *cmd.Error {
	if combustion.IsEmbeddedArtifactRegistryConfigured(ctx) {
		certsDir := filepath.Join(ctx.ImageConfigDir, combustion.K8sDir, combustion.HelmDir, combustion.CertsDir)
		ctx.HelmClient = helm.New(ctx.BuildDir, certsDir)
	}

	if ctx.ImageDefinition.Kubernetes.Version != "" {
		c, err := cache.New(rootDir)
		if err != nil {
			return &cmd.Error{
				UserMessage: "Setting up file caching failed.",
				LogMessage:  fmt.Sprintf("Initializing cache instance failed: %v", err),
			}
		}

		ctx.KubernetesScriptDownloader = kubernetes.ScriptDownloader{}
		ctx.KubernetesArtefactDownloader = kubernetes.ArtefactDownloader{
			Cache: c,
		}
	}

	return nil
}
