package commands

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/golangci/golangci-lint/pkg/commands/internal"
	"github.com/golangci/golangci-lint/pkg/logutils"
)

const envKeepTempFiles = "CUSTOM_GCL_KEEP_TEMP_FILES"

var (
	outputArgument = ""
)

type customCommand struct {
	cmd *cobra.Command

	cfg *internal.Configuration

	log logutils.Log
}

func newCustomCommand(logger logutils.Log) *customCommand {
	c := &customCommand{log: logger}

	customCmd := &cobra.Command{
		Use:          "custom",
		Short:        "Build a version of golangci-lint with custom linters",
		Args:         cobra.NoArgs,
		PreRunE:      c.preRunE,
		RunE:         c.runE,
		SilenceUsage: true,
	}

	customCmd.Flags().StringVarP(&outputArgument, "output", "o", "", color.GreenString("Path to output file"))

	c.cmd = customCmd

	return c
}

func (c *customCommand) preRunE(_ *cobra.Command, _ []string) error {
	cfg, err := internal.LoadConfiguration()
	if err != nil {
		return err
	}

	if outputArgument != "" {
		directory, name := path.Split(outputArgument)
		cfg.Destination = directory
		cfg.Name = name
	}

	err = cfg.Validate()
	if err != nil {
		return err
	}

	c.cfg = cfg

	return nil
}

func (c *customCommand) runE(cmd *cobra.Command, _ []string) error {
	tmp, err := os.MkdirTemp(os.TempDir(), "custom-gcl")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}

	defer func() {
		if os.Getenv(envKeepTempFiles) != "" {
			log.Printf("WARN: The env var %s has been detected: the temporary directory is preserved: %s", envKeepTempFiles, tmp)

			return
		}

		_ = os.RemoveAll(tmp)
	}()

	err = internal.NewBuilder(c.log, c.cfg, tmp).Build(cmd.Context())
	if err != nil {
		return fmt.Errorf("build process: %w", err)
	}

	return nil
}
