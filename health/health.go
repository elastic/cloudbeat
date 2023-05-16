package health

import (
	"errors"
	"sync"
)

var Reporter = &reporter{
	ch:     make(chan error, 1),
	errors: map[string]error{},
	mut:    sync.Mutex{},
}

type reporter struct {
	ch     chan error
	errors map[string]error
	mut    sync.Mutex
}

func (r *reporter) NewHealth(component string, err error) {
	r.mut.Lock()
	defer r.mut.Unlock()
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
