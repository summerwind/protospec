package http2

import (
	"context"
	"errors"
	"fmt"

	"github.com/summerwind/protospec/task"
	"golang.org/x/net/http2"
)

var knownSettingID = map[string]http2.SettingID{
	"SETTINGS_HEADER_TABLE_SIZE":      http2.SettingHeaderTableSize,
	"SETTINGS_ENABLE_PUSH":            http2.SettingEnablePush,
	"SETTINGS_MAX_CONCURRENT_STREAMS": http2.SettingMaxConcurrentStreams,
	"SETTINGS_INITIAL_WINDOW_SIZE":    http2.SettingInitialWindowSize,
	"SETTINGS_MAX_FRAME_SIZE":         http2.SettingMaxFrameSize,
	"SETTINGS_MAX_HEADER_LIST_SIZE":   http2.SettingMaxHeaderListSize,
}

func init() {
	task.RegisterTaskType("http2_send_settings", TaskSendSettings{})
	task.RegisterTaskType("http2_wait_settings", TaskWaitSettings{})
	task.RegisterTaskType("http2_send_ping", TaskSendPing{})
	task.RegisterTaskType("http2_wait_ping", TaskWaitPing{})
}

type Setting struct {
	ID  string
	Val uint32
}

type TaskSendSettings struct {
	Settings []Setting
}

func (t *TaskSendSettings) Run(ctx context.Context) error {
	var err error

	conn, err := getConnection(ctx)
	if err != nil {
		return err
	}

	settings := []http2.Setting{}
	for _, s := range t.Settings {
		id, ok := knownSettingID[s.ID]
		if !ok {
			return fmt.Errorf("invalid setting ID: %s", s.ID)
		}
		settings = append(settings, http2.Setting{id, s.Val})
	}

	err = conn.WriteSettings(settings...)
	if err != nil {
		return err
	}

	return nil
}

type TaskWaitSettings struct {
	Ack bool
}

func (t *TaskWaitSettings) Run(ctx context.Context) error {
	var (
		passed bool
		err    error
	)

	conn, err := getConnection(ctx)
	if err != nil {
		return err
	}

	for !conn.Closed {
		ev, err := conn.WaitEvent()
		if err != nil {
			return err
		}

		fev, ok := ev.(*FrameEvent)
		if !ok {
			return &task.Failed{
				Expect: t.expect(),
				Actual: conn.LastEvent.String(),
			}
		}

		if fev.Frame.Header().Type != http2.FrameSettings {
			continue
		}

		sf, ok := fev.Frame.(*http2.SettingsFrame)
		if !ok {
			return errors.New("invalid SETTINGS frame")
		}

		if sf.IsAck() != t.Ack {
			continue
		}

		passed = true
		break
	}

	if !passed {
		return &task.Failed{
			Expect: t.expect(),
			Actual: conn.LastEvent.String(),
		}
	}

	return nil
}

func (t *TaskWaitSettings) expect() []string {
	var flags http2.Flags

	if t.Ack {
		flags = http2.FlagSettingsAck
	}

	return []string{
		fmt.Sprintf("SETTINGS frame (length:8, flags:0x%02x, stream_id:0)", flags),
	}
}

type TaskSendPing struct {
	Ack  bool
	Data string
}

func (t *TaskSendPing) Run(ctx context.Context) error {
	var err error

	conn, err := getConnection(ctx)
	if err != nil {
		return err
	}

	data := [8]byte{}
	for i := 0; i < len(data); i++ {
		if i < len(t.Data) {
			data[i] = t.Data[i]
		}
	}

	err = conn.WritePing(t.Ack, data)
	if err != nil {
		return err
	}

	return nil
}

type TaskWaitPing struct {
	Ack  bool
	Data string
}

func (t *TaskWaitPing) Run(ctx context.Context) error {
	var (
		passed bool
		err    error
	)

	conn, err := getConnection(ctx)
	if err != nil {
		return err
	}

	for !conn.Closed {
		ev, err := conn.WaitEvent()
		if err != nil {
			return err
		}

		fev, ok := ev.(*FrameEvent)
		if !ok {
			return &task.Failed{
				Expect: t.expect(),
				Actual: conn.LastEvent.String(),
			}
		}

		if fev.Frame.Header().Type != http2.FramePing {
			continue
		}

		pf, ok := fev.Frame.(*http2.PingFrame)
		if !ok {
			return errors.New("invalid PING frame")
		}

		if pf.IsAck() != t.Ack {
			continue
		}

		if len(t.Data) > 0 {
			var data [8]byte
			copy(data[:], t.Data)

			matched := true
			for i, b := range pf.Data {
				if data[i] != b {
					matched = false
					break
				}
			}

			if !matched {
				continue
			}
		}

		passed = true
		break
	}

	if !passed {
		return &task.Failed{
			Expect: t.expect(),
			Actual: conn.LastEvent.String(),
		}
	}

	return nil
}

func (t *TaskWaitPing) expect() []string {
	var (
		flags http2.Flags
		data  [8]byte
	)

	if t.Ack {
		flags = http2.FlagPingAck
	}

	copy(data[:], t.Data)

	return []string{
		fmt.Sprintf("PING frame (length:8, flags:0x%02x, stream_id:0, opaque_data:%s)", flags, data),
	}
}
