package build

import (
	"fmt"

	"github.com/suse-edge/edge-image-builder/pkg/log"
)


func (b *Builder) Extract() error {
	log.Audit("Generating image customization components...")

	if err := b.configureCombustion(b.context); err != nil {
		log.Audit("Error configuring customization components, check the logs under the build directory for more information.")
		return fmt.Errorf("configuring combustion: %w", err)
	}


	log.Audit("Extract complete!")
	return nil
}