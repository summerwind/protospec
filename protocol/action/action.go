package action

import (
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

type ActionFailure struct {
	Reason string
}

func (e *ActionFailure) Error() string {
	return e.Reason
}

type ActionSkip struct {
	Reason string
}

func (e *ActionSkip) Error() string {
	return e.Reason
}

var (
	ErrTimeout = &ActionFailure{
		Reason: "timeout",
	}

	ErrConnectionClosed = &ActionFailure{
		Reason: "connection closed",
	}
)

func Fail(reason string) *ActionFailure {
	return &ActionFailure{
		Reason: reason,
	}
}

func Failf(format string, a ...interface{}) *ActionFailure {
	return &ActionFailure{
		Reason: fmt.Sprintf(format, a...),
	}
}

func Skip(reason string) *ActionSkip {
	return &ActionSkip{
		Reason: reason,
	}
}

func Skipf(format string, a ...interface{}) *ActionSkip {
	return &ActionSkip{
		Reason: fmt.Sprintf(format, a...),
	}
}

func IsFailure(err error) bool {
	var failure *ActionFailure
	return errors.As(err, &failure)
}

func IsSkip(err error) bool {
	var skip *ActionSkip
	return errors.As(err, &skip)
}

func IsConnectionClosed(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, io.ErrClosedPipe)
}

func IsTimeout(err error) bool {
	opErr, ok := err.(*net.OpError)
	return ok && opErr.Timeout()
}

func HandleConnectionFailure(err error) error {
	if IsConnectionClosed(err) {
		return ErrConnectionClosed
	}

	if IsTimeout(err) {
		return ErrTimeout
	}

	return err
}
