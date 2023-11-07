package pipeline

import "context"

func CollectToSlice[S ~[]T, T any](ctx context.Context, input <-chan T) (slice S, err error) {
	for item := range input {
		select {
		case <- ctx.Done():
			return nil, context.Cause(ctx)
		default:
			slice = append(slice, item)
		}
	}

	return slice, context.Cause(ctx)
}
