package contextual

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

const (
	durShort = 100 * time.Millisecond
	durLong  = durShort * 2
)

var (
	ctxBg              = context.Background()
	errPass            = fmt.Errorf("test passed")
	errFail            = fmt.Errorf("test failed")
	connPass, connFail = net.Pipe()
)

func TestCancelledContext(t *testing.T) {
	d := Dialer{&testNeverDial{}}
	ctx, cancel := context.WithCancel(ctxBg)
	cancel()
	_, err := d.DialContext(ctx, "tcp", "127.0.0.1:12345")
	if err != context.Canceled {
		t.Errorf("got: %#v want: %#v", err, context.Canceled)
	}
}

type testNeverDial struct{}

func (d *testNeverDial) Dial(network, address string) (net.Conn, error) {
	return nil, errFail
}

func TestDialContext(t *testing.T) {
	d := Dialer{&testDialContext{}}
	_, err := d.DialContext(ctxBg, "tcp", "127.0.0.1:12345")
	if err != errPass {
		t.Errorf("got: %#v want: %#v", err, errPass)
	}
}

type testDialContext struct{}

func (d *testDialContext) Dial(network, address string) (net.Conn, error) {
	return nil, errFail
}

func (d *testDialContext) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return nil, errPass
}

func TestDialTimeout(t *testing.T) {
	d := Dialer{&testDialTimeout{}}
	ctx, cancel := context.WithTimeout(ctxBg, durShort)
	defer cancel()
	_, err := d.DialContext(ctx, "tcp", "127.0.0.1:12345")
	if err != errPass {
		t.Errorf("got: %#v want: %#v", err, errPass)
	}
}

type testDialTimeout struct{}

func (d *testDialTimeout) Dial(network, address string) (net.Conn, error) {
	return nil, errFail
}

func (d *testDialTimeout) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return nil, errPass
}

func TestDialTimeoutNoDeadline(t *testing.T) {
	d := Dialer{&testDialTimeoutNoDeadline{}}
	ctx, cancel := context.WithCancel(ctxBg)
	defer cancel()
	_, err := d.DialContext(ctx, "tcp", "127.0.0.1:12345")
	if err != errPass {
		t.Errorf("got: %#v want: %#v", err, errPass)
	}
}

type testDialTimeoutNoDeadline struct{}

func (d *testDialTimeoutNoDeadline) Dial(network, address string) (net.Conn, error) {
	return nil, errPass
}

func (d *testDialTimeoutNoDeadline) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return nil, errFail
}

func TestDialSucceed(t *testing.T) {
	d := Dialer{&testDialSucceed{}}
	conn, err := d.DialContext(ctxBg, "tcp", "127.0.0.1:12345")
	if conn != connPass {
		t.Errorf("got: %#v want: %#v", conn, connPass)
	}
	if err != nil {
		t.Errorf("got: %#v want: %#v", err, nil)
	}
}

type testDialSucceed struct{}

func (d *testDialSucceed) Dial(network, address string) (net.Conn, error) {
	return connPass, nil
}

func TestDialLate(t *testing.T) {
	sentinel, _ := net.Pipe()
	d := Dialer{&testDialLate{sentinel}}
	ctx, cancel := context.WithTimeout(ctxBg, durShort)
	defer cancel()
	conn, err := d.DialContext(ctx, "tcp", "127.0.0.1:12345")
	if conn != nil {
		t.Errorf("got: %#v want: %#v", conn, nil)
	}
	if err != context.DeadlineExceeded {
		t.Errorf("got: %#v want: %#v", err, context.DeadlineExceeded)
	}
	_, err = sentinel.Write([]byte("a"))
	if err != io.ErrClosedPipe {
		t.Errorf("got: %#v want: %#v", err, io.ErrClosedPipe)
	}
}

func TestDialLateCancelled(t *testing.T) {
	sentinel, _ := net.Pipe()
	d := Dialer{&testDialLate{sentinel}}
	ctx, cancel := context.WithCancel(ctxBg)
	go func() {
		time.Sleep(durShort)
		cancel()
	}()
	conn, err := d.DialContext(ctx, "tcp", "127.0.0.1:12345")
	if conn != nil {
		t.Errorf("got: %#v want: %#v", conn, nil)
	}
	if err != context.Canceled {
		t.Errorf("got: %#v want: %#v", err, context.Canceled)
	}
	_, err = sentinel.Write([]byte("a"))
	if err != io.ErrClosedPipe {
		t.Errorf("got: %#v want: %#v", err, io.ErrClosedPipe)
	}
}

type testDialLate struct {
	sentinel net.Conn
}

func (d *testDialLate) Dial(network, address string) (net.Conn, error) {
	time.Sleep(durLong)
	return d.sentinel, nil
}
