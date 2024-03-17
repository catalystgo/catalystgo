package closer

import (
	"os"
	"time"
)

type Option func(c *closer)

func WithNoTimeout() Option {
	return func(c *closer) {
		c.noTimeout = true
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *closer) {
		c.timeout = timeout
	}
}

func WithSignals(sig ...os.Signal) Option {
	return func(c *closer) {
		c.signals = sig
	}
}
