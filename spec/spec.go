package spec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	_ "github.com/summerwind/protospec/task/http2"
	_ "github.com/summerwind/protospec/task/tcp"
)

type Spec struct {
	ID    string
	Name  string
	Tests []Test
}

type Test struct {
	Name        string
	Requirement string
	Connection  Connection
	Steps       []Step
}

//func (t *Test) Run(c *config.Config) error {
//	var (
//		conn net.Conn
//		err  error
//	)
//
//	for name, opts := range t.Connection {
//		conn, err = task.NewConnection(name, c, []byte(opts))
//		if err != nil {
//			return err
//		}
//	}
//
//	ctx := task.SetConnection(context.Background(), conn)
//
//	for _, step := range t.Steps {
//		for name, opts := range step {
//			task, err := task.NewTask(name, c, []byte(opts))
//			if err != nil {
//				return err
//			}
//
//			err = task.Run(ctx)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

type Connection map[string]json.RawMessage
type Step map[string]json.RawMessage

type Result struct {
	Passed  int
	Skipped int
	Failed  int
}

func Load(files []string) ([]*Spec, error) {
	specs := []*Spec{}

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			dirFiles, err := ioutil.ReadDir(f)
			if err != nil {
				return nil, err
			}

			subFiles := []string{}
			for _, df := range dirFiles {
				if !strings.HasSuffix(df.Name(), ".yaml") {
					subFiles = append(subFiles, filepath.Join(info.Name(), df.Name()))
				}
				if !strings.HasSuffix(df.Name(), ".yml") {
					subFiles = append(subFiles, filepath.Join(info.Name(), df.Name()))
				}
			}

			subSpecs, err := Load(subFiles)
			if err != nil {
				return nil, err
			}

			specs = append(specs, subSpecs...)
			continue
		}

		buf, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec file: %v", err)
		}

		spec := Spec{}
		err = yaml.Unmarshal(buf, &spec)
		if err != nil {
			return nil, fmt.Errorf("invalid spec: %v", err)
		}

		specs = append(specs, &spec)
	}

	return specs, nil
}
