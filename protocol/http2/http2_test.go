package http2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

func newTestConn(t *testing.T) (*Conn, net.Conn) {
	server, client := net.Pipe()

	conn, err := NewConn(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	conn.SetTimeout(100 * time.Millisecond)

	return conn, server
}

func TestInitParam(t *testing.T) {
	p := InitParam{
		Handshake: true,
		Settings: Settings{
			"SETTINGS_INITIAL_WINDOW_SIZE": 16384,
		},
		MaxFieldValueLength: 4096,
	}

	err := p.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendDataParam(t *testing.T) {
	tests := []struct {
		param SendDataParam
		err   bool
	}{
		{
			param: SendDataParam{
				Data: []string{"protospec", "protospec"},
			},
			err: false,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestSendDataFrameParam(t *testing.T) {
	tests := []struct {
		param SendDataFrameParam
		err   bool
	}{
		{
			param: SendDataFrameParam{
				StreamID:  1,
				EndStream: true,
				PadLength: 1,
				Data:      "protospec",
			},
			err: false,
		},
		{
			param: SendDataFrameParam{
				StreamID:   1,
				EndStream:  true,
				PadLength:  1,
				Data:       "protospec",
				DataLength: 10,
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestSendHeadersFrameParam(t *testing.T) {
	tests := []struct {
		param SendHeadersFrameParam
		err   bool
	}{
		{
			param: SendHeadersFrameParam{
				StreamID:   1,
				EndStream:  true,
				EndHeaders: true,
				PadLength:  1,
				HeaderFields: []Field{
					{
						Name:  "protospec",
						Value: "protospec",
					},
				},
				Priority: &Priority{
					StreamDependency: 0,
					Exclusive:        true,
					Weight:           255,
				},
			},
			err: false,
		},
		{
			param: SendHeadersFrameParam{
				StreamID:         1,
				EndStream:        true,
				EndHeaders:       true,
				PadLength:        1,
				HeaderFields:     []Field{},
				NoDefaultFields:  true,
				FillMaxFrameSize: false,
				Priority: &Priority{
					StreamDependency: 0,
					Exclusive:        true,
					Weight:           255,
				},
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestSendPriorityFrameParam(t *testing.T) {
	p := SendPriorityFrameParam{
		StreamID: 1,
		Priority: Priority{
			StreamDependency: 0,
			Exclusive:        true,
			Weight:           255,
		},
	}

	err := p.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendSettingsFrameParam(t *testing.T) {
	tests := []struct {
		param SendSettingsFrameParam
		err   bool
	}{
		{
			param: SendSettingsFrameParam{
				Ack: false,
				Settings: Settings{
					"SETTINGS_HEADER_TABLE_SIZE":      1024,
					"SETTINGS_ENABLE_PUSH":            1,
					"SETTINGS_MAX_CONCURRENT_STREAMS": 100,
					"SETTINGS_INITIAL_WINDOW_SIZE":    16384,
					"SETTINGS_MAX_FRAME_SIZE":         4096,
					"SETTINGS_MAX_HEADER_LIST_SIZE":   1024,
					"SETTINGS_UNKNOWN":                1,
				},
			},
			err: false,
		},
		{
			param: SendSettingsFrameParam{
				Ack: false,
				Settings: Settings{
					"SETTINGS_INVALID": 1,
				},
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestSendPingFrameParam(t *testing.T) {
	p := SendPingFrameParam{
		Ack:  false,
		Data: "protospec",
	}

	err := p.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendContinuationFrameParam(t *testing.T) {
	tests := []struct {
		param SendContinuationFrameParam
		err   bool
	}{
		{
			param: SendContinuationFrameParam{
				StreamID:   1,
				EndHeaders: true,
				HeaderFields: []Field{
					{
						Name:  "protospec",
						Value: "protospec",
					},
				},
			},
			err: false,
		},
		{
			param: SendContinuationFrameParam{
				StreamID:     1,
				EndHeaders:   true,
				HeaderFields: []Field{},
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestWaitHeadersFrameParam(t *testing.T) {
	p := WaitHeadersFrameParam{
		StreamID: 1,
	}

	err := p.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWaitSettingsFrameParam(t *testing.T) {
	tests := []struct {
		param WaitSettingsFrameParam
		err   bool
	}{
		{
			param: WaitSettingsFrameParam{
				Ack: false,
				Settings: Settings{
					"SETTINGS_HEADER_TABLE_SIZE":      1024,
					"SETTINGS_ENABLE_PUSH":            1,
					"SETTINGS_MAX_CONCURRENT_STREAMS": 100,
					"SETTINGS_INITIAL_WINDOW_SIZE":    16384,
					"SETTINGS_MAX_FRAME_SIZE":         4096,
					"SETTINGS_MAX_HEADER_LIST_SIZE":   1024,
					"SETTINGS_UNKNOWN":                1,
				},
			},
			err: false,
		},
		{
			param: WaitSettingsFrameParam{
				Ack: false,
				Settings: Settings{
					"SETTINGS_INVALID": 1,
				},
			},
			err: true,
		},
		{
			param: WaitSettingsFrameParam{
				Ack: true,
				Settings: Settings{
					"SETTINGS_INITIAL_WINDOW_SIZE": 16384,
				},
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestWaitPingFrameParam(t *testing.T) {
	p := WaitPingFrameParam{
		Ack:  false,
		Data: "protospec",
	}

	err := p.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWaitConnectionErrorParam(t *testing.T) {
	tests := []struct {
		param WaitConnectionErrorParam
		err   bool
	}{
		{
			param: WaitConnectionErrorParam{
				ErrorCode: []string{"PROTOCOL_ERROR"},
			},
			err: false,
		},
		{
			param: WaitConnectionErrorParam{
				ErrorCode: []string{"INVLAID_ERROR"},
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestWaitStreamErrorParam(t *testing.T) {
	tests := []struct {
		param WaitStreamErrorParam
		err   bool
	}{
		{
			param: WaitStreamErrorParam{
				StreamID:  1,
				ErrorCode: []string{"PROTOCOL_ERROR"},
			},
			err: false,
		},
		{
			param: WaitStreamErrorParam{
				StreamID:  1,
				ErrorCode: []string{"INVLAID_ERROR"},
			},
			err: true,
		},
	}

	for i, test := range tests {
		err := test.param.Validate()
		if test.err {
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] unexpected error: %v", i, err)
			}
		}
	}
}

func TestInit(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		conn, server := newTestConn(t)
		ch := make(chan error, 1)

		go func() {
			buf := make([]byte, 24)
			_, err := server.Read(buf)
			if err != nil {
				ch <- err
				return
			}

			if string(buf) != http2.ClientPreface {
				ch <- fmt.Errorf("unexpected client preface: %s", buf)
				return
			}

			framer := http2.NewFramer(server, server)
			framer.AllowIllegalWrites = true
			framer.AllowIllegalReads = true

			f, err := framer.ReadFrame()
			if err != nil {
				ch <- err
				return
			}

			sf, ok := f.(*http2.SettingsFrame)
			if !ok {
				ch <- fmt.Errorf("unexpected frame: %s", f)
				return
			}

			if sf.IsAck() {
				ch <- fmt.Errorf("unexpected SETTINGS frame: %s", f)
				return
			}

			err = framer.WriteSettingsAck()
			if err != nil {
				ch <- err
				return
			}

			err = framer.WriteSettings(http2.Setting{
				ID:  http2.SettingInitialWindowSize,
				Val: 1024,
			})
			if err != nil {
				ch <- err
				return
			}

			f, err = framer.ReadFrame()
			if err != nil {
				ch <- err
				return
			}

			sf, ok = f.(*http2.SettingsFrame)
			if !ok {
				ch <- fmt.Errorf("unexpected frame: %s", f)
				return
			}

			if !sf.IsAck() {
				ch <- fmt.Errorf("unexpected SETTINGS frame: %s", f)
				return
			}

			close(ch)
		}()

		param := InitParam{
			Handshake: true,
			Settings: Settings{
				"SETTINGS_INITIAL_WINDOW_SIZE": 1024,
			},
		}

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		err = conn.Init(buf)
		if err != nil {
			t.Errorf("init error: %v", err)
		}

		err = <-ch
		if err != nil {
			t.Errorf("server error: %v", err)
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		err = conn.Init(buf)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestRun_SendData(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []struct {
			param SendDataParam
			data  string
		}{
			{
				param: SendDataParam{
					Data: []string{"hello", "world"},
				},
				data: "helloworld",
			},
			{
				param: SendDataParam{
					Data: []string{"0x01", "0x02"},
				},
				data: "\x01\x02",
			},
		}

		for i, test := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				buf, err := ioutil.ReadAll(server)
				if err != nil {
					ch <- err
					return
				}

				if string(buf) != test.data {
					ch <- fmt.Errorf("[%d] unexpected data: %s", i, buf)
					return
				}

				close(ch)
			}()

			buf, err := json.Marshal(test.param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionSendData, buf)
			if err != nil {
				t.Errorf("[%d] run error: %v", i, err)
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionSendData, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}

func TestRun_SendDataFrame(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []SendDataFrameParam{
			{
				StreamID:  1,
				EndStream: false,
				PadLength: 10,
				Data:      "hello",
			},
			{
				StreamID:   1,
				EndStream:  false,
				PadLength:  0,
				DataLength: 10,
			},
			{
				StreamID:         1,
				EndStream:        false,
				PadLength:        0,
				FillMaxFrameSize: true,
			},
		}

		for i, param := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true

				f, err := framer.ReadFrame()
				if err != nil {
					ch <- err
					return
				}

				df, ok := f.(*http2.DataFrame)
				if !ok {
					ch <- fmt.Errorf("unexpected frame: %s", f)
					return
				}

				if df.StreamID != param.StreamID {
					ch <- fmt.Errorf("unexpected stream ID: %s", f)
					return
				}

				if df.StreamEnded() != param.EndStream {
					ch <- fmt.Errorf("unexpected END_STREAM: %s", f)
					return
				}

				if param.PadLength > 0 {
					padLen := int(df.Length) - len(df.Data()) - 1
					if uint8(padLen) != param.PadLength {
						ch <- fmt.Errorf("unexpected pad length: %s", f)
						return
					}
				}

				if param.DataLength > 0 || param.FillMaxFrameSize {
					dataLength := param.DataLength
					if param.FillMaxFrameSize {
						dataLength = conn.MaxFrameSize()
					}

					if len(df.Data()) != int(dataLength) {
						ch <- fmt.Errorf("unexpected data length: %d", len(df.Data()))
						return
					}
				} else {
					if string(df.Data()) != param.Data {
						ch <- fmt.Errorf("unexpected data: %s", df.Data())
						return
					}
				}

				close(ch)
			}()

			buf, err := json.Marshal(param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionSendDataFrame, buf)
			if err != nil {
				t.Errorf("[%d] run error: %v", i, err)
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionSendDataFrame, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}

func TestRun_SendHeadersFrame(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []struct {
			param SendHeadersFrameParam
			err   bool
		}{
			{
				param: SendHeadersFrameParam{
					StreamID:        3,
					EndStream:       false,
					EndHeaders:      false,
					PadLength:       10,
					NoDefaultFields: false,
					HeaderFields: []Field{
						{
							Name:  "test",
							Value: "test",
						},
					},
					Priority: &Priority{
						StreamDependency: 1,
						Exclusive:        true,
						Weight:           16,
					},
				},
				err: false,
			},
			{
				param: SendHeadersFrameParam{
					StreamID:        1,
					EndStream:       true,
					EndHeaders:      true,
					PadLength:       0,
					NoDefaultFields: true,
					HeaderFields: []Field{
						{
							Name:  "test",
							Value: "test",
						},
					},
				},
				err: false,
			},
			{
				param: SendHeadersFrameParam{
					StreamID:        1,
					EndStream:       true,
					EndHeaders:      true,
					PadLength:       0,
					NoDefaultFields: true,
					HeaderFields:    []Field{},
				},
				err: true,
			},
		}

		for i, test := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true

				f, err := framer.ReadFrame()
				if err != nil {
					ch <- err
					return
				}

				hf, ok := f.(*http2.HeadersFrame)
				if !ok {
					ch <- fmt.Errorf("unexpected frame: %s", f)
					return
				}

				if hf.StreamID != test.param.StreamID {
					ch <- fmt.Errorf("unexpected stream ID: %s", f)
					return
				}

				if hf.HeadersEnded() != test.param.EndHeaders {
					ch <- fmt.Errorf("unexpected END_HEADERS: %s", f)
					return
				}

				if hf.StreamEnded() != test.param.EndStream {
					ch <- fmt.Errorf("unexpected END_STREAM: %s", f)
					return
				}

				if test.param.PadLength > 0 {
					padLen := int(hf.Length) - len(hf.HeaderBlockFragment()) - 1
					if test.param.Priority != nil {
						padLen -= 5
					}

					if uint8(padLen) != test.param.PadLength {
						ch <- fmt.Errorf("unexpected pad length: %s", f)
						return
					}
				}

				if test.param.Priority != nil {
					if hf.Priority.StreamDep != test.param.Priority.StreamDependency {
						ch <- fmt.Errorf("unexpected stream dependency: %s", f)
						return
					}

					if hf.Priority.Exclusive != test.param.Priority.Exclusive {
						ch <- fmt.Errorf("unexpected exclusive: %s", f)
						return
					}

					if hf.Priority.Weight != test.param.Priority.Weight {
						ch <- fmt.Errorf("unexpected weight: %s", f)
						return
					}
				}

				var fields []Field
				decoder := hpack.NewDecoder(4096, func(f hpack.HeaderField) {
					fields = append(fields, Field(f))
				})

				if !test.param.NoDefaultFields {
					fields = conn.setDefaultHeaderFields(fields)
				}

				_, err = decoder.Write(hf.HeaderBlockFragment())
				for _, pf := range test.param.HeaderFields {
					valid := false
					for _, f := range fields {
						if pf.Name == f.Name && pf.Value == f.Value {
							valid = true
						}
					}

					if !valid {
						ch <- fmt.Errorf("unexpected header field: %s", pf.Name)
						return
					}
				}

				close(ch)
			}()

			buf, err := json.Marshal(test.param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionSendHeadersFrame, buf)
			if test.err {
				if err == nil {
					t.Errorf("[%d] unexpected nil error", i)
				}
			} else {
				if err != nil {
					t.Errorf("[%d] run error: %v", i, err)
				}

				if res != nil {
					t.Errorf("[%d] unexpected result: %v", i, res)
				}

				err = <-ch
				if err != nil {
					t.Errorf("[%d] server error: %v", i, err)
				}
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionSendHeadersFrame, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}

func TestRun_SendSettingsFrame(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []SendSettingsFrameParam{
			{
				Ack: false,
				Settings: map[string]uint32{
					"SETTINGS_HEADER_TABLE_SIZE":      4096,
					"SETTINGS_ENABLE_PUSH":            1,
					"SETTINGS_MAX_CONCURRENT_STREAMS": 100,
					"SETTINGS_INITIAL_WINDOW_SIZE":    65535,
					"SETTINGS_MAX_FRAME_SIZE":         16384,
					"SETTINGS_MAX_HEADER_LIST_SIZE":   1024,
					"SETTINGS_UNKNOWN":                1,
				},
			},
			{
				Ack: true,
				Settings: map[string]uint32{
					"SETTINGS_INITIAL_WINDOW_SIZE": 65535,
				},
			},
		}

		for i, param := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true

				f, err := framer.ReadFrame()
				if err != nil {
					ch <- err
					return
				}

				sf, ok := f.(*http2.SettingsFrame)
				if !ok {
					ch <- fmt.Errorf("unexpected frame: %s", f)
					return
				}

				if sf.IsAck() != param.Ack {
					ch <- fmt.Errorf("unexpected ack: %v", sf.IsAck())
					return
				}

				if sf.IsAck() {
					num := sf.NumSettings()
					if num != 0 {
						ch <- fmt.Errorf("unexpected number of settings: %d", num)
						return
					}
				} else {
					for key, val := range param.Settings {
						id, ok := settingID[key]
						if !ok {
							ch <- fmt.Errorf("invalid setting key: %s", key)
							return
						}

						v, ok := sf.Value(id)
						if !ok {
							ch <- fmt.Errorf("value not found: %s", key)
							return
						}

						if v != val {
							ch <- fmt.Errorf("unexpected value: key:%s, value:%d", key, v)
							return
						}
					}
				}

				close(ch)
			}()

			buf, err := json.Marshal(param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionSendSettingsFrame, buf)
			if err != nil {
				t.Errorf("[%d] run error: %v", i, err)
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		tests := []interface{}{
			"invalid",
			SendSettingsFrameParam{
				Ack: false,
				Settings: map[string]uint32{
					"SETTINGS_INVALID": 1,
				},
			},
		}

		for i, param := range tests {
			conn, _ := newTestConn(t)

			buf, err := json.Marshal(param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionSendSettingsFrame, buf)
			if err == nil {
				t.Errorf("[%d] unexpected nil error", i)
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}
		}
	})
}

func TestRun_SendPingFrame(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []SendPingFrameParam{
			{
				Ack:  false,
				Data: "01234",
			},
			{
				Ack:  false,
				Data: "",
			},
			{
				Ack:  false,
				Data: "0123456789",
			},
			{
				Ack:  true,
				Data: "01234",
			},
		}

		for i, param := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true

				f, err := framer.ReadFrame()
				if err != nil {
					ch <- err
					return
				}

				pf, ok := f.(*http2.PingFrame)
				if !ok {
					ch <- fmt.Errorf("unexpected frame: %s", f)
					return
				}

				if pf.IsAck() != param.Ack {
					ch <- fmt.Errorf("unexpected ack: %v", pf.IsAck())
					return
				}

				var data [8]byte
				copy(data[:], param.Data)

				if string(pf.Data[:]) != string(data[:]) {
					ch <- fmt.Errorf("unexpected data: %s", pf.Data)
					return
				}

				close(ch)
			}()

			buf, err := json.Marshal(param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionSendPingFrame, buf)
			if err != nil {
				t.Errorf("[%d] run error: %v", i, err)
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionSendPingFrame, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}

func TestRun_WaitHeadersFrame(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []struct {
			param    WaitHeadersFrameParam
			streamID uint32
			err      bool
		}{
			{
				param: WaitHeadersFrameParam{
					StreamID: 1,
				},
				streamID: 1,
				err:      false,
			},
			{
				param: WaitHeadersFrameParam{
					StreamID: 3,
				},
				streamID: 1,
				err:      true,
			},
		}

		for i, test := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				var err error

				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true
				framer.WritePing(false, [8]byte{'t', 'e', 's', 't'})

				encoderBuf := new(bytes.Buffer)
				encoder := hpack.NewEncoder(encoderBuf)

				encoderBuf.Reset()
				encoder.WriteField(hpack.HeaderField{Name: ":method", Value: "GET"})
				encoder.WriteField(hpack.HeaderField{Name: ":scheme", Value: "http"})
				encoder.WriteField(hpack.HeaderField{Name: ":path", Value: "/"})

				hfp := http2.HeadersFrameParam{
					StreamID:      test.streamID,
					EndStream:     false,
					EndHeaders:    true,
					BlockFragment: encoderBuf.Bytes(),
				}

				err = framer.WriteHeaders(hfp)

				if err != nil {
					ch <- err
					return
				}

				close(ch)
			}()

			buf, err := json.Marshal(test.param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionWaitHeadersFrame, buf)
			if test.err {
				if err == nil {
					t.Errorf("[%d] unexpected nil error", i)
				}
			} else {
				if err != nil {
					t.Errorf("[%d] run error: %v", i, err)
				}
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})
}

func TestRun_WaitSettingsFrame(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []struct {
			param    WaitSettingsFrameParam
			ack      bool
			settings map[string]uint32
			err      bool
		}{
			{
				param: WaitSettingsFrameParam{
					Ack: false,
				},
				ack:      false,
				settings: map[string]uint32{},
				err:      false,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: false,
					Settings: map[string]uint32{
						"SETTINGS_HEADER_TABLE_SIZE":      4096,
						"SETTINGS_ENABLE_PUSH":            1,
						"SETTINGS_MAX_CONCURRENT_STREAMS": 100,
						"SETTINGS_INITIAL_WINDOW_SIZE":    65535,
						"SETTINGS_MAX_FRAME_SIZE":         16384,
						"SETTINGS_MAX_HEADER_LIST_SIZE":   1024,
						"SETTINGS_UNKNOWN":                1,
					},
				},
				ack: false,
				settings: map[string]uint32{
					"SETTINGS_HEADER_TABLE_SIZE":      4096,
					"SETTINGS_ENABLE_PUSH":            1,
					"SETTINGS_MAX_CONCURRENT_STREAMS": 100,
					"SETTINGS_INITIAL_WINDOW_SIZE":    65535,
					"SETTINGS_MAX_FRAME_SIZE":         16384,
					"SETTINGS_MAX_HEADER_LIST_SIZE":   1024,
					"SETTINGS_UNKNOWN":                1,
				},
				err: false,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: false,
					Settings: map[string]uint32{
						"SETTINGS_HEADER_TABLE_SIZE": 4096,
					},
				},
				ack: false,
				settings: map[string]uint32{
					"SETTINGS_HEADER_TABLE_SIZE": 4096,
					"SETTINGS_ENABLE_PUSH":       1,
				},
				err: true,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: false,
					Settings: map[string]uint32{
						"SETTINGS_HEADER_TABLE_SIZE": 4096,
					},
				},
				ack: false,
				settings: map[string]uint32{
					"SETTINGS_ENABLE_PUSH": 1,
				},
				err: true,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: false,
					Settings: map[string]uint32{
						"SETTINGS_HEADER_TABLE_SIZE": 4096,
					},
				},
				ack: false,
				settings: map[string]uint32{
					"SETTINGS_HEADER_TABLE_SIZE": 1024,
				},
				err: true,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: false,
					Settings: map[string]uint32{
						"SETTINGS_HEADER_TABLE_SIZE": 4096,
					},
				},
				ack:      true,
				settings: map[string]uint32{},
				err:      true,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: true,
				},
				ack:      true,
				settings: map[string]uint32{},
				err:      false,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: true,
					Settings: map[string]uint32{
						"SETTINGS_HEADER_TABLE_SIZE":      4096,
						"SETTINGS_ENABLE_PUSH":            1,
						"SETTINGS_MAX_CONCURRENT_STREAMS": 100,
						"SETTINGS_INITIAL_WINDOW_SIZE":    65535,
						"SETTINGS_MAX_FRAME_SIZE":         16384,
						"SETTINGS_MAX_HEADER_LIST_SIZE":   1024,
						"SETTINGS_UNKNOWN":                1,
					},
				},
				ack:      true,
				settings: map[string]uint32{},
				err:      false,
			},
			{
				param: WaitSettingsFrameParam{
					Ack: true,
				},
				ack:      false,
				settings: map[string]uint32{},
				err:      true,
			},
		}

		for i, test := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				var err error

				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true
				framer.WritePing(false, [8]byte{'t', 'e', 's', 't'})

				if test.ack {
					err = framer.WriteSettingsAck()
				} else {
					var settings []http2.Setting

					for key, val := range test.settings {
						sid, ok := settingID[key]
						if !ok {
							t.Errorf("[%d] invalid setting: %s", i, key)
						}

						settings = append(settings, http2.Setting{ID: sid, Val: val})
					}

					err = framer.WriteSettings(settings...)
				}

				if err != nil {
					ch <- err
					return
				}

				close(ch)
			}()

			buf, err := json.Marshal(test.param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionWaitSettingsFrame, buf)
			if test.err {
				if err == nil {
					t.Errorf("[%d] unexpected nil error", i)
				}
			} else {
				if err != nil {
					t.Errorf("[%d] run error: %v", i, err)
				}
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionWaitSettingsFrame, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}

func TestRun_WaitPingFrame(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []struct {
			param WaitPingFrameParam
			ack   bool
			data  string
			err   bool
		}{
			{
				param: WaitPingFrameParam{
					Ack:  false,
					Data: "ping",
				},
				ack:  false,
				data: "ping",
				err:  false,
			},
			{
				param: WaitPingFrameParam{
					Ack:  true,
					Data: "ping",
				},
				ack:  false,
				data: "ping",
				err:  true,
			},
			{
				param: WaitPingFrameParam{
					Ack:  false,
					Data: "ping",
				},
				ack:  false,
				data: "",
				err:  true,
			},
		}

		for i, test := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				var (
					err  error
					data [8]byte
				)

				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true

				err = framer.WriteSettings(http2.Setting{
					ID:  http2.SettingInitialWindowSize,
					Val: 1024,
				})
				if err != nil {
					ch <- err
					return
				}

				copy(data[:], test.data)
				err = framer.WritePing(test.ack, data)

				if err != nil {
					ch <- err
					return
				}

				close(ch)
			}()

			buf, err := json.Marshal(test.param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionWaitPingFrame, buf)
			if test.err {
				if err == nil {
					t.Errorf("[%d] unexpected nil error", i)
				}
			} else {
				if err != nil {
					t.Errorf("[%d] run error: %v", i, err)
				}
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionWaitPingFrame, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}

func TestRun_WaitConnectionError(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []struct {
			param     WaitConnectionErrorParam
			errorCode string
			close     bool
			err       bool
		}{
			{
				param: WaitConnectionErrorParam{
					ErrorCode: []string{"PROTOCOL_ERROR", "INTERNAL_ERROR"},
				},
				errorCode: "PROTOCOL_ERROR",
				close:     false,
				err:       false,
			},
			{
				param: WaitConnectionErrorParam{
					ErrorCode: []string{"PROTOCOL_ERROR"},
				},
				errorCode: "NO_ERROR",
				close:     true,
				err:       false,
			},
			{
				param: WaitConnectionErrorParam{
					ErrorCode: []string{"PROTOCOL_ERROR"},
				},
				errorCode: "INTERNAL_ERROR",
				close:     false,
				err:       true,
			},
			{
				param: WaitConnectionErrorParam{
					ErrorCode: []string{"UNDEFINED_ERROR"},
				},
				errorCode: "UNDEFINED_ERROR",
				close:     false,
				err:       true,
			},
		}

		for i, test := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				var err error

				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true
				framer.WritePing(false, [8]byte{'t', 'e', 's', 't'})

				if test.close {
					err = server.Close()
				} else {
					code, ok := errorCode[test.errorCode]
					if !ok {
						close(ch)
						return
					}

					err = framer.WriteGoAway(1, code, []byte("test"))
				}

				if err != nil {
					ch <- err
					return
				}

				close(ch)
			}()

			buf, err := json.Marshal(test.param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionWaitConnectionError, buf)
			if test.err {
				if err == nil {
					t.Errorf("[%d] unexpected nil error", i)
				}
			} else {
				if err != nil {
					t.Errorf("[%d] run error: %v", i, err)
				}
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionWaitConnectionError, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}

func TestRun_WaitStreamError(t *testing.T) {
	t.Run("valid param", func(t *testing.T) {
		tests := []struct {
			param     WaitStreamErrorParam
			streamID  uint32
			errorCode string
			close     bool
			err       bool
		}{
			{
				param: WaitStreamErrorParam{
					StreamID:  1,
					ErrorCode: []string{"PROTOCOL_ERROR", "INTERNAL_ERROR"},
				},
				streamID:  1,
				errorCode: "PROTOCOL_ERROR",
				close:     false,
				err:       false,
			},
			{
				param: WaitStreamErrorParam{
					StreamID:  3,
					ErrorCode: []string{"PROTOCOL_ERROR", "INTERNAL_ERROR"},
				},
				streamID:  1,
				errorCode: "PROTOCOL_ERROR",
				close:     false,
				err:       true,
			},
			{
				param: WaitStreamErrorParam{
					StreamID:  1,
					ErrorCode: []string{"PROTOCOL_ERROR"},
				},
				streamID:  1,
				errorCode: "NO_ERROR",
				close:     true,
				err:       false,
			},
			{
				param: WaitStreamErrorParam{
					StreamID:  1,
					ErrorCode: []string{"PROTOCOL_ERROR"},
				},
				streamID:  1,
				errorCode: "INTERNAL_ERROR",
				close:     false,
				err:       true,
			},
			{
				param: WaitStreamErrorParam{
					StreamID:  1,
					ErrorCode: []string{"UNDEFINED_ERROR"},
				},
				streamID:  1,
				errorCode: "UNDEFINED_ERROR",
				close:     false,
				err:       true,
			},
		}

		for i, test := range tests {
			conn, server := newTestConn(t)
			ch := make(chan error, 1)

			go func() {
				var err error

				framer := http2.NewFramer(server, server)
				framer.AllowIllegalWrites = true
				framer.AllowIllegalReads = true
				framer.WritePing(false, [8]byte{'t', 'e', 's', 't'})

				if test.close {
					err = server.Close()
				} else {
					code, ok := errorCode[test.errorCode]
					if !ok {
						close(ch)
						return
					}

					err = framer.WriteRSTStream(test.streamID, code)
				}

				if err != nil {
					ch <- err
					return
				}

				close(ch)
			}()

			buf, err := json.Marshal(test.param)
			if err != nil {
				t.Errorf("[%d] marshal error: %v", i, err)
			}

			res, err := conn.Run(ActionWaitStreamError, buf)
			if test.err {
				if err == nil {
					t.Errorf("[%d] unexpected nil error", i)
				}
			} else {
				if err != nil {
					t.Errorf("[%d] run error: %v", i, err)
				}
			}

			if res != nil {
				t.Errorf("[%d] unexpected result: %v", i, res)
			}

			err = conn.Close()
			if err != nil {
				t.Errorf("[%d] close error: %v", i, err)
			}

			err = <-ch
			if err != nil {
				t.Errorf("[%d] server error: %v", i, err)
			}
		}
	})

	t.Run("invalid param", func(t *testing.T) {
		conn, _ := newTestConn(t)
		param := "invalid"

		buf, err := json.Marshal(param)
		if err != nil {
			t.Errorf("marshal error: %v", err)
		}

		res, err := conn.Run(ActionWaitStreamError, buf)
		if err == nil {
			t.Error("unexpected nil error")
		}

		if res != nil {
			t.Errorf("unexpected result: %v", res)
		}
	})
}
