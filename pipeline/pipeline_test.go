package pipeline

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	ctx, cancel = context.WithCancel(context.Background())
	inCh        = make(chan int)
)

func TestStep(t *testing.T) {
	type args struct {
		ctx          context.Context
		inputChannel chan int
		fn           func(context.Context, int) float64
		val          int
	}
	tests := []struct {
		name         string
		args         args
		want         float64
		shouldCancel bool
	}{
		{
			name: "Should receive value from output channel",
			args: args{
				ctx:          ctx,
				inputChannel: inCh,
				fn:           intToFloat,
				val:          1,
			},
			want:         1,
			shouldCancel: false,
		},
		{
			name: "Context is canceled - no value received",
			args: args{
				ctx:          ctx,
				inputChannel: inCh,
				fn:           intToFloat,
				val:          1,
			},
			want:         0,
			shouldCancel: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outCh := Step(tt.args.ctx, tt.args.inputChannel, tt.args.fn)
			tt.args.inputChannel <- tt.args.val

			if tt.shouldCancel {
				cancel()
			}

			var result float64
			select {
			case result = <-outCh:
			case <-ctx.Done():
				break
			}

			assert.Equal(t, tt.want, result)
		})
	}
}

func intToFloat(ctx context.Context, num int) float64 {
	return float64(num)
}
