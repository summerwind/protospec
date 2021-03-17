package command

import (
	"errors"
	"strings"
)

const (
	bundleConfigMediaType    = "application/vnd.protospec.config.v1+json"
	bundleTestLayerMediaType = "application/vnd.protospec.test.layer.v1+yaml"
	bundleRuleLayerMediaType = "application/vnd.protospec.rule.layer.v1+rego"
)

var bundleMediaTypes = []string{
	bundleConfigMediaType,
	bundleTestLayerMediaType,
	bundleRuleLayerMediaType,
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
