package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/summerwind/protospec/command"
)

var (
	_version = "dev"
	_commit  = "HEAD"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "protospec",
		Short:   "A decrelative protocol testing tool",
		Version: fmt.Sprintf("%s (%s)", _version, _commit),
	}

	rootCmd.PersistentFlags().Bool("help", false, "Display this help and exit")
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.SilenceUsage = true

	rootCmd.AddCommand(
		command.NewRunCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
