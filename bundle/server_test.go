package bundle

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	assert := assert.New(t)

	handler := NewServer()
	err := HostBundle("empty", map[string]string{})
	assert.NoError(err)

	err = HostBundle("otherBundle", map[string]string{"a.txt": "some text"})
	assert.NoError(err)

	err = HostBundle("overridenBundle", map[string]string{})
	assert.NoError(err)

	err = HostBundle("overridenBundle", map[string]string{})
	assert.NoError(err)

	server := httptest.NewServer(handler)

	var tests = []struct {
		path               string
		expectedStatusCode string
	}{
		{
			"/bundles/empty", "200 OK",
		},
		{
			"/bundles/otherBundle", "200 OK",
		},
		{
			"/bundles/otherbundle", "200 OK",
		},
		{
			"/bundles/overridenBundle", "200 OK",
		},
		{
			"/bundles/notExistBundle", "404 Not Found",
		},
	}

	for _, test := range tests {
		target := server.URL + test.path
		client := &http.Client{}
		res, err := client.Get(target)

		assert.NoError(err)
		assert.Equal(test.expectedStatusCode, res.Status)
	}
}
