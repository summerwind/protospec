package tcp

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/summerwind/protospec/task"
)

func init() {
	task.RegisterTaskType("tcp_send_data", TaskSendData{})
}

type TaskSendData struct {
	Data interface{}
}

func (t *TaskSendData) Run(ctx context.Context) error {
	var (
		data []string
		err  error
	)

	conn, err := task.GetConnection(ctx)
	if err != nil {
		return err
	}

	switch raw := t.Data.(type) {
	case string:
		data = []string{raw}
	case []string:
		data = raw
	default:
		return fmt.Errorf("data must be string or string array: %v", data)
	}

	for _, d := range data {
		var buf []byte

		if strings.HasPrefix(d, "0x") {
			buf, err = hex.DecodeString(d[2:])
			if err != nil {
				return err
			}
		} else {
			buf = []byte(d)
		}
		_, err := conn.Write(buf)
		if err != nil {
			return err
		}
	}

	return nil
}
