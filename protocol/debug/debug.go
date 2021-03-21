package debug

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/summerwind/protospec/protocol/action"
)

const (
	ProtocolType = "debug"

	ActionFail            = "debug.fail"
	ActionSkip            = "debug.skip"
	ActionConnectionClose = "debug.connection_close"
	ActionTimeout         = "debug.timeout"
	ActionError           = "debug.error"
)

type Conn struct {
	net.Conn
}

func NewConn(conn net.Conn) (*Conn, error) {
	return &Conn{conn}, nil
}

func (conn *Conn) Init(param []byte) error {
	return nil
}

func (conn *Conn) Run(action string, param []byte) (interface{}, error) {
	switch action {
	case ActionFail:
		return conn.fail(param)
	case ActionSkip:
		return conn.skip(param)
	case ActionConnectionClose:
		return conn.connectionClose(param)
	case ActionTimeout:
		return conn.timeout(param)
	case ActionError:
		return conn.error(param)
	default:
		return nil, fmt.Errorf("invalid action: %s", action)
	}
}

func (conn *Conn) Close() error {
	return nil
}

func (conn *Conn) SetMode(server bool) {
	// Nothing to do.
}

func (conn *Conn) SetTimeout(timeout time.Duration) {
	// Nothing to do.
}

func (conn *Conn) fail(param []byte) (interface{}, error) {
	return nil, action.Fail("this action is always failed")
}

func (conn *Conn) skip(param []byte) (interface{}, error) {
	return nil, action.Skip("this action is always skipped")
}

func (conn *Conn) connectionClose(param []byte) (interface{}, error) {
	return nil, action.ErrConnectionClosed
}

func (conn *Conn) timeout(param []byte) (interface{}, error) {
	return nil, action.ErrTimeout
}

func (conn *Conn) error(param []byte) (interface{}, error) {
	return nil, errors.New("this action returns an error")
}
