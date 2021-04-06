package command

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewPullCommand() *cobra.Command {
	var spec string

	var cmd = cobra.Command{
		Use:   "pull <bundle>",
		Short: "Pull spec bundles from an OCI registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundleName := args[0]
			return pull(bundleName, spec)
		},
	}

	pflag := cmd.Flags()
	pflag.StringVarP(&spec, "spec", "s", "spec", "Path to save the spec bundle to")

	return &cmd
}

func pull(bundleName, bundlePath string) error {
	if err := validateBundleName(bundleName); err != nil {
		return err
	}

	bundleName = normalizeBundleName(bundleName)

	store := content.NewFileStore(bundlePath)
	defer store.Close()

	ctx := context.Background()
	logrus.SetLevel(logrus.FatalLevel)

	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("auth client error: %w", err)
	}

	resolver, err := client.Resolver(ctx, http.DefaultClient, false)
	if err != nil {
		return fmt.Errorf("resolver error: %w", err)
	}

	opts := []oras.PullOpt{
		oras.WithAllowedMediaTypes(bundleMediaTypes),
	}

	info, _, err := oras.Pull(ctx, resolver, bundleName, store, opts...)
	if err != nil {
		return fmt.Errorf("failed to pull spec bundle: %w", err)
	}

	fmt.Printf("Digest: %s\n", info.Digest)
	fmt.Println(bundleName)

	return nil
}
