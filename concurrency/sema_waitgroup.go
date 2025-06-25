package concurrency

import (
	"sync"
)

type SemaWaitGroup struct {
	wg   *sync.WaitGroup
	sema chan struct{}
}

func NewSemaWaitGroup(concurrency int) *SemaWaitGroup {
	if concurrency == 0 {
		concurrency = 1
	}
	return &SemaWaitGroup{
		wg:   new(sync.WaitGroup),
		sema: make(chan struct{}, concurrency),
	}
}

func (g *SemaWaitGroup) Do(f func()) {
	g.sema <- struct{}{}
	g.wg.Add(1)

	go func() {
		defer func() {
			<-g.sema
			g.wg.Done()
		}()

		f()
	}()
}

func (g *SemaWaitGroup) Wait() {
	g.wg.Wait()
}

// Close is not necessary, will be deprecated in future
func (g *SemaWaitGroup) Close() {
	close(g.sema)
}
