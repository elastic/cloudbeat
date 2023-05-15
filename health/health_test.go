package health

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHealth(t *testing.T) {
	r := &reporter{
		ch:     make(chan error, 1),
		errors: make(map[string]error),
	}

	events := []struct {
		component string
		err       error
		wantErr   bool
	}{
		{
			component: "component1",
			err:       nil,
			wantErr:   false,
		},
		{
			component: "component1",
			err:       errors.New("component1 went wrong"),
			wantErr:   true,
		},
		{
			component: "component2",
			err:       errors.New("component2 went wrong"),
			wantErr:   true,
		},
		{
			component: "component2",
			err:       nil,
			wantErr:   true,
		},
		{
			component: "component1",
			err:       nil,
			wantErr:   false,
		},
	}

	for _, e := range events {
		r.NewHealth(e.component, e.err)
		err := <-r.ch
		if e.wantErr {
			assert.Error(t, err)
			fmt.Println(err)
		} else {
			assert.NoError(t, err)
		}
	}
}
