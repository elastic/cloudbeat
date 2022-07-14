package layerfs

import (
	"fmt"
)

type layerFsError struct {
	name string
	text string
}

func newError(text string, name string) error {
	return &layerFsError{
		text: text,
		name: name,
	}
}

func (l *layerFsError) Error() string {
	return fmt.Sprintf("go-layerfs: %s: %s", l.text, l.name)
}
