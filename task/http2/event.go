package http2

import (
	"fmt"

	"golang.org/x/net/http2"
)

type Event interface {
	String() string
}

type FrameEvent struct {
	Frame http2.Frame
}

func (ev *FrameEvent) String() string {
	switch f := ev.Frame.(type) {
	case *http2.PingFrame:
		header := f.Header()
		return fmt.Sprintf(
			"%s frame (length:%d, flags:0x%02x, stream_id:%d, opaque_data:%v)",
			header.Type,
			header.Length,
			header.Flags,
			header.StreamID,
			f.Data,
		)
	default:
		header := ev.Frame.Header()
		return fmt.Sprintf(
			"%s frame (length:%d, flags:0x%02x, stream_id:%d)",
			header.Type,
			header.Length,
			header.Flags,
			header.StreamID,
		)
	}
}

type ConnectionCloseEvent struct{}

func (ev *ConnectionCloseEvent) String() string {
	return "Connection closed"
}

type TimeoutEvent struct{}

func (ev *TimeoutEvent) String() string {
	return "Timeout"
}
