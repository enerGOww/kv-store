package closer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Func func(ctx context.Context) error

var Closer = closer{isClosing: false}

type closer struct {
	mu        sync.Mutex
	funcs     []Func
	isClosing bool
}

func (c *closer) Add(f Func) {
	if c.isClosing {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f)
}

func (c *closer) Close(ctx context.Context) error {
	c.mu.Lock()
	c.isClosing = true
	c.mu.Unlock()

	closingErr := make([]string, 0, len(c.funcs))
	wg := sync.WaitGroup{}
	wg.Add(len(c.funcs))

	for _, f := range c.funcs {
		go func(f Func) {
			if err := f(ctx); err != nil {
				c.mu.Lock()
				closingErr = append(closingErr, fmt.Sprintf("[!] %v", err.Error()))
				c.mu.Unlock()
			}
			wg.Done()
		}(f)
	}

	completing := make(chan struct{})
	go func() {
		wg.Wait()
		completing <- struct{}{}
	}()

	select {
	case <-completing:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %v", ctx.Err())
	}

	if len(closingErr) > 0 {
		return fmt.Errorf(
			"shutdown finished with error(s): \n%s",
			strings.Join(closingErr, "\n"),
		)
	}

	return nil
}
