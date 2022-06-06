package pipeline

import "context"

func Step[In any, Out any](ctx context.Context, inputChannel chan In, fn func(context.Context, In) Out) chan Out {
	outputCh := make(chan Out)

	go func() {
		defer close(outputCh)

		for s := range inputChannel {
			select {
			case <-ctx.Done():
				break
			default:
			}

			go func(ctx context.Context, s In) {
				outputCh <- fn(ctx, s)
			}(ctx, s)
		}
	}()

	return outputCh
}
