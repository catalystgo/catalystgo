package closer

import (
	"os"
	"time"
)

type Option func(c *closer)

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
