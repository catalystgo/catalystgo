package closer

import (
	"errors"
	"sort"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKeySlice(t *testing.T) {
	keys := keySlice{
		NormalOrder,
		LowOrder,
		HighOrder,
	}

	sort.Sort(keys)

	exp := keySlice{
		HighOrder,
		NormalOrder,
		LowOrder,
	}

	require.Equal(t, exp, keys)
}

func TestAdd(t *testing.T) {
	c := (New()).(*closer)

	var (
		n  = int(1e6)
		wg = sync.WaitGroup{}
	)

	wg.Add(n)
	for range n {
		go func() {
			c.Add(func() error { return nil })
			wg.Done()
		}()
	}

	wg.Wait()

	require.Len(t, c.keys, 1)
	require.Len(t, c.closers[NormalOrder], n)
}

func TestOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  []Option
		check func(t *testing.T, c *closer)
	}{
		{
			name: "new_without_opts",
			opts: []Option{},
			check: func(t *testing.T, c *closer) {
				ctx, _ := c.getContext()
				cancelTime, ok := ctx.Deadline()

				require.True(t, ok)
				require.InDelta(t, float64(time.Now().Add(30*time.Second).Second()), float64(cancelTime.Second()), float64(1*time.Millisecond))
			},
		},
		{
			name: "new_with_close_timeout",
			opts: []Option{WithTimeout(5 * time.Second)},
			check: func(t *testing.T, c *closer) {
				ctx, _ := c.getContext()
				cancelTime, ok := ctx.Deadline()

				require.True(t, ok)
				require.InDelta(t, float64(time.Now().Add(5*time.Second).Second()), float64(cancelTime.Second()), float64(1*time.Millisecond))

			},
		},
		{
			name: "new_with_no_timeout",
			opts: []Option{WithTimeout(5 * time.Second), WithNoTimeout()},
			check: func(t *testing.T, c *closer) {
				ctx, _ := c.getContext()
				cancelTime, ok := ctx.Deadline()

				require.False(t, ok)
				require.InDelta(t, float64(time.Now().Add(5*time.Second).Second()), cancelTime.Second(), float64(1*time.Millisecond))
			},
		},
		{
			name: "new_with_signals",
			opts: []Option{WithSignals(syscall.SIGTERM)},
			check: func(t *testing.T, c *closer) {
				require.Len(t, c.signals, 1)
			},
		},
		{
			name: "new_with_signals",
			opts: []Option{WithSignals(syscall.SIGTERM)},
			check: func(t *testing.T, c *closer) {
				require.Len(t, c.signals, 1)
			},
		},
		{
			name: "new_and_close",
			opts: []Option{},
			check: func(t *testing.T, c *closer) {
				order := make([]int, 3)
				m := map[Order][]func() error{
					LowOrder: {func() error {
						order[2] = 3
						return errors.New("some_error")
					}},
					NormalOrder: {func() error {
						order[1] = 2
						return errors.New("some_error")
					}},
					HighOrder: {func() error {
						order[0] = 1
						return errors.New("some_error")
					}},
				}

				for order, functions := range m {
					c.AddByOrder(order, functions...)
				}

				go c.CloseAll()
				c.Wait()

				require.Equal(t, []int{1, 2, 3}, order)
			},
		},
		{
			name: "new_and_close_with_time_out",
			opts: []Option{WithTimeout(1 * time.Second)},
			check: func(t *testing.T, c *closer) {
				order := make([]int, 2)
				m := map[Order][]func() error{
					LowOrder: {func() error {
						time.Sleep(2 * time.Second)
						return errors.New("some_error")
					}},
					NormalOrder: {func() error {
						order[1] = 2
						return errors.New("some_error")
					}},
					HighOrder: {func() error {
						order[0] = 1
						return errors.New("some_error")
					}},
				}

				for order, functions := range m {
					c.AddByOrder(order, functions...)
				}

				go c.CloseAll()
				c.Wait()

				require.Equal(t, []int{1, 2}, order)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := New(tt.opts...).(*closer)
			tt.check(t, c)
		})
	}
}
