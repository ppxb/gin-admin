package logger

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type HookExecutor interface {
	Exec(extra map[string]string, b []byte) error
	Close() error
}

type hookOptions struct {
	maxJobs    int
	maxWorkers int
	extra      map[string]string
}

type HookOption func(*hookOptions)

func WithMaxJobs(jobs int) HookOption {
	return func(options *hookOptions) {
		options.maxJobs = jobs
	}
}

func WithMaxWorkers(workers int) HookOption {
	return func(options *hookOptions) {
		options.maxWorkers = workers
	}
}

func WithExtra(extra map[string]string) HookOption {
	return func(options *hookOptions) {
		options.extra = extra
	}
}

type Hook struct {
	opts   *hookOptions
	q      chan []byte
	wg     *sync.WaitGroup
	exec   HookExecutor
	closed int32
}

func NewHook(exec HookExecutor, options ...HookOption) *Hook {
	opts := &hookOptions{
		maxJobs:    1024,
		maxWorkers: 2,
	}
	for _, o := range options {
		o(opts)
	}

	wg := new(sync.WaitGroup)
	wg.Add(opts.maxWorkers)
	h := &Hook{
		opts: opts,
		q:    make(chan []byte, opts.maxJobs),
		wg:   wg,
		exec: exec,
	}
	h.dispatch()
	return h
}

func (h *Hook) dispatch() {
	for i := 0; i < h.opts.maxWorkers; i++ {
		go func() {
			defer func() {
				h.wg.Done()
				if r := recover(); r != nil {
					fmt.Println("recovered from panic in logger hook: ", r)
				}
			}()

			for data := range h.q {
				err := h.exec.Exec(h.opts.extra, data)
				if err != nil {
					fmt.Println("failed to write entry: ", err.Error())
				}
			}
		}()
	}
}

func (h *Hook) Write(p []byte) (int, error) {
	if atomic.LoadInt32(&h.closed) == 1 || len(h.q) == h.opts.maxJobs {
		return len(p), nil
	}

	data := make([]byte, len(p))
	copy(data, p)
	select {
	case h.q <- data:
	default:
		fmt.Println("too many jobs, waiting for queue to be empty, discard")
	}
	return len(p), nil
}

func (h *Hook) Flush() {
	if atomic.CompareAndSwapInt32(&h.closed, 0, 1) {
		close(h.q)
		h.wg.Wait()
		if err := h.exec.Close(); err != nil {
			fmt.Println("failed to close logger hook:", err.Error())
		}
	}
}
