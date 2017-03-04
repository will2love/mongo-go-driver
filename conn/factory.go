package conn

import (
	"context"

	"github.com/10gen/mongo-go-driver/internal"
)

// Factory creates a connection.
type Factory func(context.Context) (Connection, error)

// DialerFactory returns a Factory that uses a dialer.
func DialerFactory(dialer Dialer, endpoint Endpoint, opts ...Option) Factory {
	return func(ctx context.Context) (Connection, error) {
		return dialer(ctx, endpoint, opts...)
	}
}

// LimitedFactory returns a Factory that is constrained by a resource
// limit.
func LimitedFactory(max uint64, factory Factory) Factory {
	permits := internal.NewSemaphore(max)
	return func(ctx context.Context) (Connection, error) {

		err := permits.Wait(ctx)
		if err != nil {
			return nil, err
		}

		c, err := factory(ctx)
		if err != nil {
			permits.Release()
			return nil, err
		}
		return &limitedFactoryConn{c, permits}, nil
	}
}

type limitedFactoryConn struct {
	Connection
	permits *internal.Semaphore
}

func (c *limitedFactoryConn) Close() error {
	c.permits.Release()
	return c.Connection.Close()
}
