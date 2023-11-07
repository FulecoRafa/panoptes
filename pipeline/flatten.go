package pipeline

import "context"

func Flatten[S ~[]T, T any](ctx context.Context, input <-chan S) (ret <-chan T) {
	ch := make(chan T)
	ret = ch
	go func() {
		defer close(ch)
		for slice := range input {
			for _, item := range slice {
				select {
				case <- ctx.Done():
					return
				case ch <- item:
				}
			}
		}
	}()
	return
}
