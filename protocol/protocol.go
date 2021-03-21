package protocol

import (
	"time"
)

type Conn interface {
	Init(param []byte) error
	Run(action string, param []byte) (interface{}, error)
	Close() error
	SetMode(server bool)
	SetTimeout(timeout time.Duration)
}
