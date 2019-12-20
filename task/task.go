package task

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"

	"github.com/summerwind/protospec/config"
)

var (
	ErrFailed  = errors.New("failed")
	ErrSkipped = errors.New("skipped")
)

// taskTypes holds the map of registered task types.
var taskTypes = map[string]reflect.Type{}

type Task interface {
	Run(context.Context) error
}

// RegisterTaskType registers a task type.
func RegisterTaskType(name string, t interface{}) {
	taskTypes[name] = reflect.TypeOf(t)
}

func NewTask(name string, c *config.Config, opts []byte) (Task, error) {
	var err error

	t, ok := taskTypes[name]
	if !ok {
		return nil, fmt.Errorf("invalid task type: %s", name)
	}

	raw := reflect.New(t).Interface()

	err = yaml.Unmarshal(opts, raw)
	if err != nil {
		return nil, err
	}

	task, ok := raw.(Task)
	if !ok {
		return nil, errors.New("unexpected task type")
	}

	return task, nil
}
