package closer

import (
	"context"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	"github.com/catalystgo/logger/logger"
)

type Closer interface {
	Add(f ...func() error)
	AddByOrder(o Order, f ...func() error)
	Wait()
	CloseAll()
}

type Order int8

// Notice the gap between values, You can use it to add custom order in between.
// Close sequence is as follows HighOrder => LowOrder
const (
	LowOrder    Order = -100
	NormalOrder Order = 0
	HighOrder   Order = 100
)

type keySlice []Order

func (s keySlice) Len() int           { return len(s) }
func (s keySlice) Less(i, j int) bool { return s[i] > s[j] }
func (s keySlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

var globalOrderCloser = New()

type closer struct {
	done chan struct{}
	mu   sync.Mutex
	once sync.Once

	keys    keySlice
	closers map[Order][]func() error

	signals []os.Signal
	timeout time.Duration
}

func New(opts ...Option) Closer {
	c := &closer{
		done:    make(chan struct{}),
		closers: make(map[Order][]func() error),
	}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.signals) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, c.signals...)
			<-ch // wait for signal
			signal.Stop(ch)
			c.CloseAll()
		}()
	}
	return c
}

// Add adds a function to the globalCloser with the default order `NormalOrder`.
func Add(f ...func() error) {
	globalOrderCloser.AddByOrder(NormalOrder, f...)
}

// AddByOrder adds a function to the globalCloser with the specified order.
func AddByOrder(o Order, f ...func() error) {
	globalOrderCloser.AddByOrder(o, f...)
}

// Wait waits for all functions in the closer to complete after calling `CloseAll`.
func Wait() {
	globalOrderCloser.Wait()
}

// CloseAll closes all functions in the globalCloser.
func CloseAll() {
	globalOrderCloser.CloseAll()
}

// Add adds a function to the closer with the default order `NormalOrder`.
func (c *closer) Add(f ...func() error) {
	c.AddByOrder(NormalOrder, f...)
}

// AddByOrder adds a function to the closer with the specified order.
func (c *closer) AddByOrder(o Order, f ...func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.closers[o]; !ok {
		c.closers[o] = f
		c.keys = append(c.keys, o)
	} else {
		c.closers[o] = append(c.closers[o], f...)
	}
}

// Wait waits for all functions in the closer to complete after calling `CloseAll`.
func (c *closer) Wait() {
	<-c.done
}

// CloseAll closes all functions in the closer.
// Order on closing functions is `HighOrder` > `NormalOrder` > `LowOrder`.
func (c *closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		closers := c.closers
		c.closers = make(map[Order][]func() error)
		c.mu.Unlock()

		c.closeAll(closers)
	})
}

func (c *closer) closeAll(closerByOrder map[Order][]func() error) {
	ctx, cancel := c.getContext()
	defer cancel()

	wg := &sync.WaitGroup{}
	sort.Sort(c.keys)

	var (
		closers []func() error
		errs    chan error
	)

	for _, k := range c.keys {
		closers = closerByOrder[k]
		errs = make(chan error)

		wg.Add(len(closers))
		for _, f := range closers {
			go func(f func() error) {
				errs <- f()
				wg.Done()
			}(f)
		}

		go func() {
			wg.Wait()
			close(errs)
		}()

	outerloop:
		for {
			select {
			case err, ok := <-errs:
				// If channel got closed then
				// continue on the first level loop
				if !ok {
					break outerloop
				}
				if err != nil {
					logger.Warnf(ctx, "error closing: %v", err)
				}
			case <-ctx.Done():
				// If the context got canceled or deadlineExceeded just return
				logger.Warnf(ctx, "error closing (ctx): %v", ctx.Err())
				return
			}
		}
	}
}

// getContext get closer's context depending on the passed options.
func (c *closer) getContext() (ctx context.Context, close func()) {
	if c.timeout == 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), c.timeout)
}
