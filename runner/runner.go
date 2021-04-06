package runner

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/summerwind/protospec/log"
	"github.com/summerwind/protospec/protocol"
	"github.com/summerwind/protospec/protocol/action"
	"github.com/summerwind/protospec/protocol/debug"
	"github.com/summerwind/protospec/protocol/http2"
	"github.com/summerwind/protospec/spec"
)

type Config struct {
	SpecPath string
	Host     string
	Port     uint32
	Tests    []string
	Insecure bool
	Strict   bool
	Timeout  uint32
	Verbose  bool
}

func (c *Config) Validate() error {
	if _, err := os.Stat(c.SpecPath); os.IsNotExist(err) {
		return fmt.Errorf("spec bundle directory does not exist: %s", c.SpecPath)
	}
	if c.Port == 0 {
		return errors.New("target port must be specified")
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

	target := (len(r.config.Tests) != 0)
	targetSpecs := map[string]bool{}
	targetTests := map[string]bool{}

	for _, test := range r.config.Tests {
		parts := strings.Split(test, "/")
		if len(parts) < 2 {
			return fmt.Errorf("invalid test ID: %s", test)
		}

		targetSpecs[strings.Join(parts[:len(parts)-1], "/")] = true
		targetTests[test] = true
	}

	bundle, err := spec.Load(r.config.SpecPath)
	if err != nil {
		return err
	}

	for _, spec := range bundle.Specs {
		if _, ok := targetSpecs[spec.ID]; target && !ok {
			continue
		}

		log.SpecName(spec.ID, spec.Name)

		for i, test := range spec.Tests {
			id := i + 1

			if !r.config.Strict && test.Optional {
				continue
			}

			if _, ok := targetTests[fmt.Sprintf("%s/%d", spec.ID, id)]; target && !ok {
				continue
			}

			if !r.config.Verbose {
				log.TestName(id, test.Name)
			}

			conn, err := r.NewConn(test)
			if err != nil {
				return err
			}

			param := test.Protocol.Param
			if len(param) == 0 {
				param = []byte("{}")
			}

			if err := conn.Init(param); err != nil {
				return err
			}

			for _, step := range test.Steps {
				param = step.Param
				if len(param) == 0 {
					param = []byte("{}")
				}

				_, err = conn.Run(step.Action, param)
				if err != nil {
					break
				}
			}

			if err != nil {
				switch {
				case action.IsFailure(err):
					log.Fail(id, test.Name, err.Error())
					failedCount += 1
				case action.IsSkip(err):
					log.Skip(id, test.Name, err.Error())
					skippedCount += 1
				default:
					log.Error(id, test.Name)
					return err
				}
			} else {
				log.Pass(id, test.Name)
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

	addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)
	transport, err = net.Dial(test.Transport, addr)
	if err != nil {
		return nil, fmt.Errorf("transport error: %w", err)
	}

	switch test.Protocol.Type {
	case http2.ProtocolType:
		conn, err = http2.NewConn(transport)
	case debug.ProtocolType:
		conn, err = debug.NewConn(transport)
	default:
		err = fmt.Errorf("invalid protocol: %s", test.Protocol.Type)
	}

	conn.SetTimeout(time.Duration(r.config.Timeout) * time.Second)

	return conn, err
}
