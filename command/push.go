package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewPushCommand() *cobra.Command {
	var spec string

	var cmd = cobra.Command{
		Use:   "push <bundle>",
		Short: "Push spec bundles to an OCI registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundleName := args[0]
			return push(bundleName, spec)
		},
	}

	pflag := cmd.Flags()
	pflag.StringVarP(&spec, "spec", "s", "spec", "Path of the spec bundle push to")

	return &cmd
}

func push(bundleName, bundlePath string) error {
	if err := validateBundleName(bundleName); err != nil {
		return err
	}

	bundleName = normalizeBundleName(bundleName)

	tests, rules, err := loadBundleFiles(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load spec files: %w", err)
	}

	store := content.NewMemoryStore()

	var layers []ocispec.Descriptor
	for path, contents := range tests {
		layers = append(layers, store.Add(path, bundleTestLayerMediaType, contents))
	}
	for path, contents := range rules {
		layers = append(layers, store.Add(path, bundleRuleLayerMediaType, contents))
	}

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

	opts := []oras.PushOpt{
		oras.WithConfigMediaType(bundleConfigMediaType),
	}

	manifest, err := oras.Push(ctx, resolver, bundleName, store, layers, opts...)
	if err != nil {
		return fmt.Errorf("failed to push spec bundle: %w", err)
	}

	fmt.Printf("Digest: %s\n", manifest.Digest)
	fmt.Println(bundleName)

	return nil
}

func loadBundleFiles(dir string) (map[string][]byte, map[string][]byte, error) {
	tests := map[string][]byte{}
	rules := map[string][]byte{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		var fileType string

		if info.IsDir() {
			return nil
		}

		switch filepath.Ext(path) {
		case ".yml":
			fileType = "test"
		case ".yaml":
			fileType = "test"
		case ".rego":
			fileType = "rule"
		}

		if fileType == "" {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if fileType == "test" {
			tests[relPath] = content
		} else {
			rules[relPath] = content
		}

		return nil
	})

	return tests, rules, err
}
