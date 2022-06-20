package pipeline

import (
	"context"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	log = logp.NewLogger("cloudbeat_config_test_suite")
)

func TestStep(t *testing.T) {
	type args struct {
		inputChannel chan int
		fn           func(context.Context, int) (float64, error)
		val          int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Should receive value from output channel",
			args: args{
				inputChannel: make(chan int),
				fn:           func(context context.Context, i int) (float64, error) { return float64(i), nil },
				val:          1,
			},
			want: 1,
		},
		{
			name: "Pipeline function returns error - no value received",
			args: args{
				inputChannel: make(chan int),
				fn:           func(context context.Context, i int) (float64, error) { return 0, errors.New("") },
				val:          1,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outCh := Step(log, tt.args.inputChannel, tt.args.fn)
			tt.args.inputChannel <- tt.args.val
			results := testhelper.CollectResources(outCh)

			assert.Equal(t, tt.want, len(results))
		})
	}
}
