package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/summerwind/protospec/spec"
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

	store := content.NewMemoryStore()
	layers, err := loadBundleFiles(store, bundlePath)
	if err != nil {
		return err
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

func loadBundleFiles(store *content.Memorystore, p string) ([]ocispec.Descriptor, error) {
	var layers []ocispec.Descriptor

	bundle, err := spec.Load(p)
	if err != nil {
		return nil, err
	}

	bundlePath := filepath.Dir(bundle.Manifest.Path)

	rp, err := filepath.Rel(bundlePath, bundle.Manifest.Path)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(bundle.Manifest.Path)
	if err != nil {
		return nil, err
	}

	layers = append(layers, store.Add(rp, bundleManifestLayerMediaType, content))

	for _, sp := range bundle.Manifest.SenderSpecs {
		specPath := filepath.Join(bundlePath, sp)

		if !strings.HasPrefix(specPath, bundlePath) {
			return nil, fmt.Errorf("the spec file is located outside the bundle directory: %s", specPath)
		}

		content, err := ioutil.ReadFile(filepath.Join(bundlePath, sp))
		if err != nil {
			return nil, err
		}

		layers = append(layers, store.Add(sp, bundleSpecLayerMediaType, content))
	}

	return layers, nil
}
