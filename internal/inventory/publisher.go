package inventory

import (
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/samber/lo"
	"time"
)

type assetInventoryPublisher struct {
	logger *logp.Logger
}

func newPublisher(logger *logp.Logger) assetInventoryPublisher {
	return assetInventoryPublisher{
		logger: logger,
	}
}

func (p assetInventoryPublisher) publish(buffer []Asset) {
	events := lo.Map(buffer, func(a Asset, _ int) beat.Event {
		return beat.Event{
			Meta:      mapstr.M{libevents.FieldMetaIndex: generateIndex(a)},
			Timestamp: time.Now(),
			Fields: mapstr.M{
				"asset": a,
			},
		}
	})

	// todo figure how to publish to elasticsearch

	for _, e := range events {
		p.logger.Infof("Publishing asset event %v", e)
	}
}

func generateIndex(a Asset) string {
	return fmt.Sprintf("asset_inventory_%s_%s_%s_%s", a.Category, a.SubCategory, a.Type, a.SubStype)
}
