package command

import (
	"github.com/spf13/cobra"
)

func NewRunCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run <addr>",
		Short: "Run tests against the target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}

	return cmd
}

func run() error {
	return nil
}
