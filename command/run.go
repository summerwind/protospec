package command

import (
	"github.com/spf13/cobra"
	"github.com/summerwind/protospec/endpoint"
)

// NewRunCommand returns the 'run' command.
func NewRunCommand() *cobra.Command {
	var c endpoint.ClientConfig

	var cmd = &cobra.Command{
		Use:   "run <addr>",
		Short: "Run tests against the target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c.Path = args[0]
			return run(c)
		},
	}

	pflag := cmd.Flags()
	pflag.StringVarP(&c.TargetAddr, "target-addr", "t", "", "Target address to be used for testing")
	pflag.StringSliceVarP(&c.Run, "run", "r", []string{}, "Test IDs to run")
	pflag.StringVar(&c.Transport, "transport", "", "Transport to be used for testing")
	pflag.Uint32VarP(&c.Timeout, "timeout", "o", 3, "Time seconds to test timeout")
	pflag.BoolVar(&c.Verbose, "verbose", false, "Output verbose log")

	return cmd
}

// run executes the client test with specified tests.
func run(c endpoint.ClientConfig) error {
	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}
