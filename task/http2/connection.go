package http2

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"

	"github.com/summerwind/protospec/config"
	"github.com/summerwind/protospec/task"
)

const (
	// DefaultWindowSize is the value of default connection window size.
	DefaultWindowSize = 65535
	// DefaultFrameSize is the value of default frame size.
	DefaultFrameSize = 16384
)

func init() {
	task.RegisterConnectionType("http2", Conn{})
}

type Conn struct {
	net.Conn
	http2.Framer

	Timeout   time.Duration
	Handshake bool `json:"handshake"`
	Closed    bool

	WindowUpdate   bool
	WindowSize     map[uint32]int
	RemoteSettings map[http2.SettingID]uint32

	LastEvent Event

	encoder *hpack.Encoder
	decoder *hpack.Decoder

	debugFramer  *http2.Framer
	debugHandler func(string)
}

func getConnection(ctx context.Context) (*Conn, error) {
	raw, err := task.GetConnection(ctx)
	if err != nil {
		return nil, err
	}

	conn, ok := raw.(*Conn)
	if !ok {
		return nil, errors.New("invalid HTTP/2 connection in the context")
	}

	return conn, nil
}

func (conn *Conn) Connect(c *config.Config) error {
	var (
		err        error
		writer     io.Writer
		encoderBuf bytes.Buffer
		//debugBuf   bytes.Buffer
	)

	if c.TLS {
		dialer := &net.Dialer{}
		dialer.Timeout = c.Timeout

		tlsConfig, err := c.TLSConfig()
		if err != nil {
			return err
		}

		if tlsConfig.NextProtos == nil {
			tlsConfig.NextProtos = append(tlsConfig.NextProtos, http2.NextProtoTLS)
		}

		tlsConn, err := tls.DialWithDialer(dialer, "tcp", c.Addr, tlsConfig)
		if err != nil {
			return err
		}

		cs := tlsConn.ConnectionState()
		if !cs.NegotiatedProtocolIsMutual {
			return errors.New("protocol negotiation failed")
		}

		conn.Conn = tlsConn
	} else {
		conn.Conn, err = net.DialTimeout("tcp", c.Addr, c.Timeout)
		if err != nil {
			return err
		}
	}

	writer = conn.Conn

	if c.Verbose {
		debugReader, debugWriter := io.Pipe()
		writer = io.MultiWriter(conn.Conn, debugWriter)

		//debugFramer := http2.NewFramer(nil, &debugBuf)
		debugFramer := http2.NewFramer(nil, debugReader)
		debugFramer.AllowIllegalReads = true

		go func() {
			for !conn.Closed {
				f, _ := debugFramer.ReadFrame()
				if f != nil {
					ev := FrameEvent{f}
					conn.writeLog(true, ev.String())
				}
			}
		}()
	}

	//framer := http2.NewFramer(io.MultiWriter(conn.Conn, &debugBuf), conn.Conn)
	framer := http2.NewFramer(writer, conn.Conn)
	framer.AllowIllegalWrites = true
	framer.AllowIllegalReads = true
	conn.Framer = *framer

	conn.encoder = hpack.NewEncoder(&encoderBuf)
	conn.decoder = hpack.NewDecoder(4096, func(f hpack.HeaderField) {})

	conn.RemoteSettings = map[http2.SettingID]uint32{}
	conn.Timeout = c.Timeout

	conn.WindowUpdate = true
	conn.WindowSize = map[uint32]int{0: DefaultWindowSize}

	if conn.Handshake {
		err = conn.handshake()
		if err != nil {
			return err
		}
	}

	return nil
}

func (conn *Conn) Write(b []byte) (int, error) {
	conn.writeLog(true, fmt.Sprintf("Raw (length:%d, data:%s...)", len(b), b[:8]))
	return conn.Conn.Write(b)
}

func (conn *Conn) Close() error {
	conn.Closed = true
	return conn.Conn.Close()
}

func (conn *Conn) HandleDebug(handler func(string)) {
	conn.debugHandler = handler
}

func (conn *Conn) WaitEvent() (Event, error) {
	t := time.Now().Add(conn.Timeout)
	conn.SetReadDeadline(t)

	f, err := conn.ReadFrame()
	if err != nil {
		if err == io.EOF {
			conn.Closed = true
			conn.LastEvent = &ConnectionCloseEvent{}
			return conn.LastEvent, nil
		}

		opErr, ok := err.(*net.OpError)
		if ok {
			if opErr.Err == syscall.ECONNRESET {
				conn.Closed = true
				conn.LastEvent = &ConnectionCloseEvent{}
				return conn.LastEvent, nil
			}

			if opErr.Timeout() {
				conn.Closed = true
				if conn.LastEvent == nil {
					conn.LastEvent = &TimeoutEvent{}
				}
				return conn.LastEvent, nil
			}
		}

		return nil, err
	}

	switch frame := f.(type) {
	case *http2.SettingsFrame:
		conn.updateRemoteSettings(frame)
	case *http2.DataFrame:
		conn.updateWindowSize(frame)
	}

	conn.LastEvent = &FrameEvent{f}
	conn.writeLog(false, conn.LastEvent.String())

	return conn.LastEvent, nil
}

func (conn *Conn) handshake() error {
	done := make(chan error)

	_, err := conn.Write([]byte(http2.ClientPreface))
	if err != nil {
		return err
	}

	go func() {
		local := false
		remote := false

		setting := http2.Setting{
			ID:  http2.SettingInitialWindowSize,
			Val: DefaultWindowSize,
		}
		conn.WriteSettings(setting)

		for !(local && remote) {
			ev, err := conn.WaitEvent()
			if err != nil {
				done <- err
				return
			}

			switch event := ev.(type) {
			case *ConnectionCloseEvent:
				done <- task.ErrClosed
				return
			case *TimeoutEvent:
				done <- task.ErrTimeout
				return
			case *FrameEvent:
				sf, ok := event.Frame.(*http2.SettingsFrame)
				if !ok {
					continue
				}

				if sf.IsAck() {
					local = true
				} else {
					remote = true
					conn.WriteSettingsAck()
				}
			}
		}

		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-time.After(conn.Timeout):
		return task.ErrTimeout
	}

	return nil
}

func (conn *Conn) updateRemoteSettings(sf *http2.SettingsFrame) {
	if sf.IsAck() {
		return
	}

	sf.ForeachSetting(func(setting http2.Setting) error {
		conn.RemoteSettings[setting.ID] = setting.Val
		return nil
	})
}

func (conn *Conn) updateWindowSize(df *http2.DataFrame) {
	if !conn.WindowUpdate {
		return
	}

	len := int(df.Header().Length)
	streamID := df.Header().StreamID

	_, ok := conn.WindowSize[streamID]
	if !ok {
		conn.WindowSize[streamID] = DefaultWindowSize
	}

	conn.WindowSize[streamID] -= len
	if conn.WindowSize[streamID] <= 0 {
		incr := DefaultWindowSize + (conn.WindowSize[streamID] * -1)
		conn.WriteWindowUpdate(streamID, uint32(incr))
		conn.WindowSize[streamID] += incr
	}

	conn.WindowSize[0] -= len
	if conn.WindowSize[0] <= 0 {
		incr := DefaultWindowSize + (conn.WindowSize[0] * -1)
		conn.WriteWindowUpdate(0, uint32(incr))
		conn.WindowSize[0] += incr
	}
}

func (conn *Conn) writeLog(send bool, str string) {
	var prefix string

	if conn.debugHandler == nil {
		return
	}

	if send {
		prefix = "S"
	} else {
		prefix = "R"
	}

	conn.debugHandler(fmt.Sprintf("[%s] %s", prefix, str))
}
