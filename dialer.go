package contextual

import (
	"context"
	"net"
	"time"
)

// SimpleDialer represents a simple net.Conn Dialer interface,
// without context.Context support.
type SimpleDialer interface {
	Dial(network, address string) (net.Conn, error)
}

// Dialer wraps and provides DialContext to an embedded Dialer.
type Dialer struct {
	SimpleDialer
}

// DialContext dials a network connection using the embedded SimpleDialer.
// If the underlying SimpleDialer implements DialContext, use it directly.
// It returns any errors, except if the Dial returns after ctx expires.
// Supplied ctx must be non-nil.
func (d Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if cd, ok := d.SimpleDialer.(contextDialer); ok {
		return cd.DialContext(ctx, network, address)
	}
	c := make(chan conAir)
	go d.dial(ctx, network, address, c)
	select {
	case res := <-c:
		return res.conn, res.err
	case <-ctx.Done():
	}
	return nil, ctx.Err()
}

func (d Dialer) dial(ctx context.Context, network, address string, c chan<- conAir) {
	var res conAir
	td, hasDialTimeout := d.SimpleDialer.(timeoutDialer)
	deadline, hasDeadline := ctx.Deadline()
	if hasDialTimeout && hasDeadline {
		res.conn, res.err = td.DialTimeout(network, address, time.Until(deadline))
	} else {
		res.conn, res.err = d.Dial(network, address)
	}
	select {
	case c <- res: // Caller is still waiting for an answer
		return
	case <-ctx.Done(): // Caller has gone away or ctx closed
	}
	// If we make it this far, close the connection
	if res.err == nil && res.conn != nil {
		res.conn.Close()
	}
}

// https://imgur.com/IpEwSZ6
type conAir struct {
	conn net.Conn
	err  error
}

type contextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type timeoutDialer interface {
	DialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
}

var (
	_ SimpleDialer  = &net.Dialer{}
	_ contextDialer = &Dialer{}
)
