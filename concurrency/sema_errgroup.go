package concurrency

import (
	"context"
	"sync"
)

type SemaErrGroup struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error

	sema chan struct{}
}

func NewSemaErrGroupWithContext(ctx context.Context, concurrency int) (*SemaErrGroup, context.Context) {
	if concurrency == 0 {
		concurrency = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	return &SemaErrGroup{
		cancel: cancel,
		sema:   make(chan struct{}, concurrency),
	}, ctx
}

func NewSemaErrGroup(concurrency int) *SemaErrGroup {
	if concurrency == 0 {
		concurrency = 1
	}
	return &SemaErrGroup{
		sema: make(chan struct{}, concurrency),
	}
}

func (g *SemaErrGroup) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

func (g *SemaErrGroup) Do(f func() error) {
	g.sema <- struct{}{}

	g.wg.Add(1)

	go func() {
		defer func() {
			<-g.sema
			g.wg.Done()
		}()

		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

// Close is not necessary, will be deprecated in future
func (g *SemaErrGroup) Close() {
	close(g.sema)
}
