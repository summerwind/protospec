package command

import (
	"errors"
	"strings"
)

const (
	bundleConfigMediaType        = "application/vnd.protospec.config.v1+json"
	bundleManifestLayerMediaType = "application/vnd.protospec.manifest.layer.v1+yaml"
	bundleSpecLayerMediaType     = "application/vnd.protospec.spec.layer.v1+yaml"
)

var bundleMediaTypes = []string{
	bundleConfigMediaType,
	bundleManifestLayerMediaType,
	bundleSpecLayerMediaType,
}

func validateBundleName(name string) error {
	if !strings.Contains(name, "/") {
		return errors.New("bundle name must contain the repository name")
	}

	return nil
}

func normalizeBundleName(name string) string {
	parts := strings.Split(name, "/")
	if !strings.Contains(parts[len(parts)-1], ":") {
		name = name + ":latest"
	}

	return name
}
