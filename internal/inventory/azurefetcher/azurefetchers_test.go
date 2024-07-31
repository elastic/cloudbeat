package azurefetcher

import (
	"context"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/stretchr/testify/assert"
)

func collectResourcesAndMatch(t *testing.T, fetcher inventory.AssetFetcher, expected []inventory.AssetEvent) {
	t.Helper()

	ch := make(chan inventory.AssetEvent)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	go func() {
		fetcher.Fetch(ctx, ch)
	}()

	received := make([]inventory.AssetEvent, 0, len(expected))
	for len(expected) != len(received) {
		select {
		case <-ctx.Done():
			assert.ElementsMatch(t, expected, received)
			return
		case event := <-ch:
			received = append(received, event)
		}
	}

	assert.ElementsMatch(t, expected, received)
}
