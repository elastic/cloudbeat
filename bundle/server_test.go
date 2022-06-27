package bundle

import (
	"context"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"net/http"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	assert := assert.New(t)

	handler := NewServer()

	ctx := context.Background()

	eksBundle := CISEksBundle()
	err := HostBundle("eks", eksBundle, ctx)
	assert.NoError(err)

	kubernetesBundle := CISKubernetesBundle()
	err = HostBundle("kubernetes", kubernetesBundle, ctx)
	assert.NoError(err)

	emptyBundle := Bundle{fs: fstest.MapFS{}}
	err = HostBundle("empty", emptyBundle, ctx)
	assert.NoError(err)

	otherBundleFS := Bundle{fs: fstest.MapFS{
		"a.txt": {
			Data: []byte("some text"),
		},
	}}
	err = HostBundle("otherBundle", otherBundleFS, ctx)
	assert.NoError(err)

	err = HostBundle("overridenBundle", emptyBundle, ctx)
	assert.NoError(err)

	err = HostBundle("overridenBundle", emptyBundle, ctx)
	assert.NoError(err)

	server := httptest.NewServer(handler)

	var tests = []struct {
		path               string
		expectedStatusCode string
	}{
		{
			"/bundles/eks", "200 OK",
		},
		{
			"/bundles/kubernetes", "200 OK",
		},
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
