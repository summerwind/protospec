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

type Spec struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Ref   string `json:"ref,omitempty"`
	Tests []Test `json:"tests"`
}

type Test struct {
	Name      string       `json:"name"`
	Desc      string       `json:"desc,omitempty"`
	Ref       string       `json:"ref,omitempty"`
	Transport string       `json:"transport,omitempty"`
	Protocol  TestProtocol `json:"protocol"`
	Steps     []TestStep   `json:"steps"`
}

type TestProtocol struct {
	Type   string          `json:"type"`
	Params json.RawMessage `json:"params"`
}

type TestStep struct {
	Action string          `json:"action"`
	Params json.RawMessage `json:"params"`
	Rule   TestRule        `json:"rule,omitempty"`
}

type TestRule struct {
	Path string `json:"path"`
}

func Load(dir string) ([]Spec, error) {
	var specs []Spec

	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(p)
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}

		buf, err := ioutil.ReadFile(p)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		var spec Spec
		err = yaml.Unmarshal(buf, &spec)
		if err != nil {
			return fmt.Errorf("failed to parse spec: %w", err)
		}

		specs = append(specs, spec)

		return nil
	})

	return specs, err
}
