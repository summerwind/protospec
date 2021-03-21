package command

import (
	"github.com/spf13/cobra"
	"github.com/summerwind/protospec/runner"
)

func NewRunCommand() *cobra.Command {
	var c runner.Config

	var cmd = &cobra.Command{
		Use:   "run <addr>",
		Short: "Run a spec test against the target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c.Addr = args[0]
			return run(c)
		},
	}

	pflag := cmd.Flags()
	pflag.StringVarP(&c.SpecPath, "spec", "s", "spec", "Path of spec tests directory to run")
	pflag.BoolVarP(&c.Insecure, "insecure", "k", false, "Don't verify server's certificate")
	pflag.BoolVar(&c.Strict, "strict", false, "Run all test cases including optional test cases")
	pflag.Uint32VarP(&c.Timeout, "timeout", "o", 3, "Time seconds to test timeout")
	pflag.BoolVar(&c.Verbose, "verbose", false, "Output verbose log")

	return cmd
}

func run(c runner.Config) error {
	if err := c.Validate(); err != nil {
		return err
	}

	r := runner.NewRunner(c)
	return r.Run()
}
