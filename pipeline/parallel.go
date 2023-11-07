package pipeline

import (
	"context"
	"errors"
	"log"
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
				log.Default().Printf("Parallel processing: %v -> %v ; %v", thisInput, thisOutput, err)
				if err != nil {
					if errors.Is(err, SkipError) {
						log.Default().Printf("[SKIP] Parallel processing: %v -> %v ; %v", thisInput, thisOutput, err)
						continue
					}
					log.Default().Printf("[ERR] Parallel processing: %v -> %v ; %v", thisInput, thisOutput, err)
					cancelCauseFunc(err)
				}
				select {
				case <- ctx.Done():
					log.Default().Printf("Parallel processing: context is done")
					return
				case ch<-thisOutput:
					log.Default().Printf("Parallel processing: sending %v", thisOutput)
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
