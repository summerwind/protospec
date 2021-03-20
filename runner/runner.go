package runner

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/summerwind/protospec/log"
	"github.com/summerwind/protospec/protocol"
	"github.com/summerwind/protospec/protocol/http2"
	"github.com/summerwind/protospec/spec"
)

type Config struct {
	Addr     string
	SpecPath string
	Insecure bool
	Timeout  uint32
	Verbose  bool
}

func (c *Config) Validate() error {
	if _, err := os.Stat(c.SpecPath); os.IsNotExist(err) {
		return fmt.Errorf("spec directory does not exist: %s", c.SpecPath)
	}

	return nil
}

type Runner struct {
	config Config
}

func NewRunner(c Config) *Runner {
	return &Runner{
		config: c,
	}
}

func (r *Runner) Run() error {
	var (
		err          error
		testCount    int
		passedCount  int
		failedCount  int
		skippedCount int
	)

	log.Verbose = r.config.Verbose

	specs, err := spec.Load(r.config.SpecPath)
	if err != nil {
		return err
	}

	for _, spec := range specs {
		log.SpecName(spec.ID, spec.Name)

		for i, test := range spec.Tests {
			id := i + 1

			if !r.config.Verbose {
				log.TestName(id, test.Name)
			}

			conn, err := r.NewConn(test)
			if err != nil {
				return err
			}

			if err := conn.Init(test.Protocol.Param); err != nil {
				return err
			}

			for _, step := range test.Steps {
				_, err = conn.Run(step.Action, step.Param)
				if err != nil {
					if protocol.IsActionError(err) {
						break
					} else {
						return err
					}
				}
			}

			if err != nil {
				if protocol.IsSkipped(err) {
					log.Skipped(id, test.Name, err.Error())
					skippedCount += 1
				} else {
					log.Failed(id, test.Name, err.Error())
					failedCount += 1
				}
			} else {
				log.Passed(id, test.Name)
				passedCount += 1
			}

			if err := conn.Close(); err != nil {
				return err
			}

			testCount += 1
		}

		log.BlankLine()
	}

	log.Summary(testCount, passedCount, skippedCount, failedCount)

	return nil
}

func (r *Runner) NewConn(test spec.Test) (protocol.Conn, error) {
	var (
		transport net.Conn
		conn      protocol.Conn
		err       error
	)

	switch test.Transport {
	case spec.TransportTCP, spec.TransportUDP:
		// OK
	case "":
		test.Transport = spec.TransportTCP
	default:
		return nil, fmt.Errorf("invalid transport: '%s'", test.Transport)
	}

	transport, err = net.Dial(test.Transport, r.config.Addr)
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	switch test.Protocol.Type {
	case http2.ProtocolType:
		conn, err = http2.NewConn(transport)
	default:
		err = fmt.Errorf("invalid protocol: %s", test.Protocol.Type)
	}

	conn.SetTimeout(time.Duration(r.config.Timeout) * time.Second)

	return conn, err
}
