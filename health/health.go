package health

import (
	"errors"
)

var Reporter = &reporter{
	ch:     make(chan error, 1),
	errors: make(map[string]error),
}

type reporter struct {
	ch     chan error
	errors map[string]error
}

func (r *reporter) NewHealth(component string, err error) {
	r.errors[component] = err
	list := make([]error, 0, len(r.errors))
	for _, err := range r.errors {
		if err != nil {
			list = append(list, err)
		}
	}

	r.ch <- errors.Join(list...)
}

func (r *reporter) Channel() <-chan error {
	return r.ch
}

func (r *reporter) Close() {
	close(r.ch)
}
