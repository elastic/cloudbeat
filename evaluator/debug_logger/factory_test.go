package dlogger

import (
	"testing"

	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/stretchr/testify/assert"
)

func TestFactoryNew(t *testing.T) {
	f := Factory{}
	manager, err := plugins.New([]byte{}, "test", inmem.New())
	assert.NoError(t, err)

	p := f.New(manager, config{})
	assert.NotNil(t, p)
}

func TestFactoryValidate(t *testing.T) {
	f := Factory{}
	manager, err := plugins.New([]byte{}, "test", inmem.New())
	assert.NoError(t, err)

	cfg, err := f.Validate(manager, []byte{})
	assert.NoError(t, err)
	assert.IsType(t, config{}, cfg)
}
