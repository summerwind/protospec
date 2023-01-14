package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
)

// Endpoint types.
const (
	EndpointTypeClient = "client"
	EndpointTypeServer = "server"
)

// Endporint transports.
const (
	EndpointTransportTCP = "tcp"
	EndpointTransportUDP = "udp"
)

// TestStep is a step of test case.
type TestStep struct {
	Action string          `json:"action"`
	Param  json.RawMessage `json:"param"`
}

// Validate validates test step data and returns error if data is invalid.
func (ts *TestStep) Validate() error {
	if ts.Action == "" {
		return errors.New("action must be specified")
	}

	return nil
}

// Test represents a single test case of protocol.
type Test struct {
	Name  string     `json:"name"`
	Desc  string     `json:"desc,omitempty"`
	Ref   string     `json:"ref,omitempty"`
	Steps []TestStep `json:"steps"`
}

// Validate validates test data and returns error if data is invalid.
func (t *Test) Validate() error {
	if t.Name == "" {
		return errors.New("name must be specified")
	}

	if len(t.Steps) == 0 {
		return errors.New("at least one step must be included in the test")
	}

	for i, s := range t.Steps {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("step #%d: %w", i, err)
		}
	}

	return nil
}

// Endpoint contains endpoint type and transport protocol.
type Endpoint struct {
	Type      string          `json:"type"`
	Transport string          `json:"transport"`
	Param     json.RawMessage `json:"param"`
}

// Validate validates endpoint data and returns error if data is invalid.
func (e *Endpoint) Validate() error {
	if e.Type == "" {
		return errors.New("endpoint type must be specified")
	}

	if e.Transport == "" {
		return errors.New("endpoint transport must be specified")
	}

	return nil
}

// Spec is data containing multiple tests and metadata.
type Spec struct {
	Version  string   `json:"version"`
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Ref      string   `json:"ref,omitempty"`
	Path     string   `json:"-"`
	Endpoint Endpoint `json:"endpoint"`
	Tests    []Test   `json:"tests"`
}

// Validate validates spec data and returns error if data is invalid.
func (s *Spec) Validate() error {
	if err := s.Endpoint.Validate(); err != nil {
		return err
	}

	if len(s.Tests) == 0 {
		return errors.New("at least one test must be included in the spec")
	}

	for i, t := range s.Tests {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("test #%d: %w", i, err)
		}
	}

	return nil
}

// Load loads spec file from specified path and returns a spec.
func Load(p string) (Spec, error) {
	var spec Spec

	buf, err := os.ReadFile(p)
	if err != nil {
		return spec, fmt.Errorf("failed to read file: %s - %w", p, err)
	}

	err = yaml.Unmarshal(buf, &spec)
	if err != nil {
		return spec, fmt.Errorf("failed to parse spec: %s - %w", p, err)
	}

	spec.Path = p

	return spec, nil
}
