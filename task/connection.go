package task

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"

	"gopkg.in/yaml.v2"

	"github.com/summerwind/protospec/config"
)

// connTypes holds the map of registered connection types.
var connTypes = map[string]reflect.Type{}

var connContextKey = "conn"

var (
	ErrClosed  = errors.New("connection closed")
	ErrTimeout = errors.New("timeout")
)

type Conn interface {
	net.Conn
	Connect(*config.Config) error
	HandleDebug(func(string))
}

// RegisterConnectionType registers a connection type.
func RegisterConnectionType(name string, t interface{}) {
	connTypes[name] = reflect.TypeOf(t)
}

func NewConnection(name string, opts []byte) (Conn, error) {
	var err error

	t, ok := connTypes[name]
	if !ok {
		return nil, fmt.Errorf("invalid connection type: %s", name)
	}

	raw := reflect.New(t).Interface()

	err = yaml.Unmarshal(opts, raw)
	if err != nil {
		return nil, err
	}

	conn, ok := raw.(Conn)
	if !ok {
		return nil, errors.New("unexpected connection type")
	}

	return conn, nil
}

func SetConnection(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, connContextKey, conn)
}

func GetConnection(ctx context.Context) (net.Conn, error) {
	raw := ctx.Value(connContextKey)
	if raw == nil {
		return nil, errors.New("no connection in the context")
	}

	conn, ok := raw.(net.Conn)
	if !ok {
		return nil, errors.New("invalid connection")
	}

	return conn, nil
}
