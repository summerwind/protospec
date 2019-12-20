package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/summerwind/protospec/config"
	"github.com/summerwind/protospec/spec"
)

var (
	_version = "0.0.0"
	_commit  = "Unknown"

	specs    []string
	timeout  int
	tls      bool
	insecure bool
	strict   bool
	verbose  bool
	dryrun   bool
	version  bool
	help     bool
)

func main() {
	cmd := &cobra.Command{
		Use:           "protospec",
		Short:         "Conformance testing tool for internet protocols",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          run,
	}

	flags := cmd.Flags()
	flags.StringArrayVarP(&specs, "file", "f", []string{}, "Path of spec file")
	flags.IntVarP(&timeout, "timeout", "t", 2, "Time seconds to test timeout")
	flags.BoolVarP(&tls, "tls", "T", false, "Connect over TLS")
	flags.BoolVarP(&insecure, "insecure", "k", false, "Don't verify server's certificate")
	flags.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose log")
	flags.BoolVar(&dryrun, "dryrun", false, "Display only test names")
	flags.BoolVar(&version, "version", false, "Display version information and exit")
	flags.BoolVar(&help, "help", false, "Display this help and exit")

	err := cmd.Execute()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	if version {
		fmt.Printf("Version: %s (%s)\n", _version, _commit)
		return nil
	}

	if help || len(args) < 1 {
		return cmd.Usage()
	}

	c := config.Config{
		Addr:     args[0],
		Specs:    specs,
		Timeout:  time.Duration(timeout) * time.Second,
		TLS:      tls,
		Insecure: insecure,
		Verbose:  verbose,
		DryRun:   dryrun,
	}

	err := c.Validate()
	if err != nil {
		return err
	}

	r := spec.NewRunner(&c)
	err = r.Run()
	if err != nil {
		return err
	}

	return nil
}
