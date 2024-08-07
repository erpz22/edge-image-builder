package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)


func NewExtractCommand(action func(*cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:      "extract",
		Usage:     "Extract new image",
		UsageText: fmt.Sprintf("%s extract [OPTIONS]", appName),
		Action:    action,
		Flags: []cli.Flag{
			DefinitionFileFlag,
			ConfigDirFlag,
			&cli.StringFlag{
				Name:        "extract-dir",
				Usage:       "Full path to the directory to store extract artifacts",
				Destination: &BuildArgs.RootBuildDir,
			},
		},
	}
}
