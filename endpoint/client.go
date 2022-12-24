package endpoint

import (
	"fmt"
	"os"
)

// ClientConfig represents a configuration of client endpoint.
type ClientConfig struct {
	Path       string
	TargetAddr string
	Run        []string
	Transport  string
	Timeout    uint32
	Verbose    bool
}

// Validate validates configuration values.
func (c *ClientConfig) Validate() error {
	if _, err := os.Stat(c.Path); os.IsNotExist(err) {
		return fmt.Errorf("test file does not exist: %s", c.Path)
	}

	return nil
}
