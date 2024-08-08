package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/log"
)

func (b *Builder) Extract() error {
	log.Audit("Generating image customization components...")

	if err := b.configureCombustion(b.context); err != nil {
		log.Audit("Error configuring customization components, check the logs under the extract directory for more information.")
		return fmt.Errorf("configuring combustion: %w", err)
	}

	log.Audit("Extract complete!")
	return nil
}

func SetupExtractArtifactsDirectory(buildDir string) (artefactsDir string, err error) {
	artefactsDir = filepath.Join(buildDir, "artefacts")
	if err = os.MkdirAll(artefactsDir, os.ModePerm); err != nil {
		return  "", fmt.Errorf("creating an artefacts directory: %w", err)
	}

	return  artefactsDir, nil
}