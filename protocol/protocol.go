package protocol

import (
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"
)

const (
	CodeFailed  = "failed"
	CodeSkipped = "skipped"

	ReasonTimeout = "timeout"
	ReasonClosed  = "connection closed"
)

var (
	ErrFailed = &ActionError{
		Code: CodeFailed,
	}

	ErrSkipped = &ActionError{
		Code: CodeSkipped,
	}

	ErrTimeout = &ActionError{
		Code:   CodeFailed,
		Reason: ReasonTimeout,
	}

	ErrConnectionClosed = &ActionError{
		Code:   CodeFailed,
		Reason: ReasonClosed,
	}
)

type ActionError struct {
	Code   string
	Reason string
}

func (e *ActionError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

func IsActionError(err error) bool {
	_, ok := err.(*ActionError)
	return ok
}

func IsFailed(err error) bool {
	actionErr, ok := err.(*ActionError)
	return ok && (actionErr.Code == CodeFailed)
}

func IsSkipped(err error) bool {
	actionErr, ok := err.(*ActionError)
	return ok && (actionErr.Code == CodeSkipped)
}

func IsConnectionClosed(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, io.ErrClosedPipe)
}

func IsTimeout(err error) bool {
	opErr, ok := err.(*net.OpError)
	return ok && opErr.Timeout()
}

func HandleConnectionError(err error) error {
	if IsConnectionClosed(err) {
		return ErrConnectionClosed
	}

	if IsTimeout(err) {
		return ErrTimeout
	}

	return err
}

type Conn interface {
	Init(params []byte) error
	Run(action string, params []byte) (interface{}, error)
	SetMode(server bool)
	SetTimeout(timeout time.Duration)
	Close() error
}
