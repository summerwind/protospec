package spec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

const (
	TransportTCP = "tcp"
	TransportUDP = "udp"
)

type Manifest struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Ref           string   `json:"ref,omitempty"`
	SenderSpecs   []string `json:"sender_specs"`
	ReceiverSpecs []string `json:"receiver_specs"`
}

type Bundle struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Ref   string `json:"ref,omitempty"`
	Specs []Spec `json:"specs"`
}

type Spec struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Ref   string `json:"ref,omitempty"`
	Path  string `json:"-"`
	Tests []Test `json:"tests"`
}

type Test struct {
	Name      string       `json:"name"`
	Desc      string       `json:"desc,omitempty"`
	Ref       string       `json:"ref,omitempty"`
	Optional  bool         `json:"optional,omitempty"`
	Transport string       `json:"transport,omitempty"`
	Protocol  TestProtocol `json:"protocol"`
	Steps     []TestStep   `json:"steps"`
}

type TestProtocol struct {
	Type  string          `json:"type"`
	Param json.RawMessage `json:"param"`
}

type TestStep struct {
	Action string          `json:"action"`
	Param  json.RawMessage `json:"param"`
	Rule   TestRule        `json:"rule,omitempty"`
}

type TestRule struct {
	Path string `json:"path"`
}

func Load(p string) (*Bundle, error) {
	stat, err := os.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %s - %w", p, err)
	}

	if stat.IsDir() {
		var found bool

		manifestPath := []string{
			filepath.Join(p, "manifest.yaml"),
			filepath.Join(p, "manifest.yml"),
		}

		for _, mp := range manifestPath {
			stat, err := os.Stat(mp)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, fmt.Errorf("failed to read: %s - %w", mp, err)
			}

			if !stat.IsDir() {
				p = mp
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("no manifest file found in %s", p)
		}
	}

	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %s - %w", p, err)
	}

	var manifest Manifest
	err = yaml.Unmarshal(buf, &manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %s - %w", p, err)
	}

	bundle := Bundle{
		ID:   manifest.ID,
		Name: manifest.Name,
		Ref:  manifest.Ref,
	}

	for _, sp := range manifest.SenderSpecs {
		specPath := filepath.Join(filepath.Dir(p), sp)

		buf, err := ioutil.ReadFile(specPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %s - %w", sp, err)
		}

		var spec Spec
		err = yaml.Unmarshal(buf, &spec)
		if err != nil {
			return nil, fmt.Errorf("failed to parse spec: %s - %w", sp, err)
		}

		spec.Path = specPath
		bundle.Specs = append(bundle.Specs, spec)
	}

	return &bundle, nil
}
