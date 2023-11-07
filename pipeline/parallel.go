package pipeline

import (
	"context"
	"sync"
)

type ExecutorFunc[I, O any] func(ctx context.Context, input I) (output O, err error)

func RunInParallel[I, O any](ctx context.Context, cancelCauseFunc context.CancelCauseFunc, workerN int, inputChan <-chan I, executorFunc ExecutorFunc[I,O]) (outputChan <-chan O) {
	ch := make(chan O)
	outputChan = ch
	var wg sync.WaitGroup
	wg.Add(workerN)
	for i := 0; i < workerN; i++ {
		go func() {
			defer wg.Done()
			for thisInput := range inputChan {
				thisOutput, err := executorFunc(ctx, thisInput)
				if err != nil {
					cancelCauseFunc(err)
				}
				select {
				case ch<-thisOutput:
				case <- ctx.Done():
					return
				}
			}
		}()
	}
	go func() {
		defer close(ch)
		wg.Wait()
	}()
	return
}
