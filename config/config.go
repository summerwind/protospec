package config

import (
	"crypto/tls"
	"errors"
	"time"
)

type Config struct {
	Addr     string
	Specs    []string
	Timeout  time.Duration
	TLS      bool
	Insecure bool
	Verbose  bool
	DryRun   bool
}

func (c *Config) Validate() error {
	if c.Addr == "" {
		return errors.New("Target address must be specified")
	}

	if len(c.Specs) == 0 {
		return errors.New("at least one spec file must be specified")
	}

	return nil
}

// TLSConfig returns a tls.Config based on the configuration.
func (c *Config) TLSConfig() (*tls.Config, error) {
	if !c.TLS {
		return nil, nil
	}

	tc := tls.Config{
		InsecureSkipVerify: c.Insecure,
	}

	return &tc, nil
}
