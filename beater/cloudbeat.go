package beater

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/cloudbeat/evaluator"

	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/cloudbeat/config"
	_ "github.com/elastic/cloudbeat/processor" // Add cloudbeat default processors.
	"github.com/elastic/cloudbeat/resources"
	"github.com/elastic/cloudbeat/transformer"

	"github.com/gofrs/uuid"
)

// cloudbeat configuration.
type cloudbeat struct {
	ctx    context.Context
	cancel context.CancelFunc

	config      config.Config
	client      beat.Client
	data        *resources.Data
	evaluator   evaluator.Evaluator
	transformer transformer.Transformer
}

const (
	cycleStatusStart = "start"
	cycleStatusEnd   = "end"
	processesDir     = "/hostfs"
)

// New creates an instance of cloudbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		cancel()
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	logp.Info("Config initiated.")

	fetchersRegistry, err := InitRegistry(ctx, c)
	if err != nil {
		cancel()
		return nil, err
	}

	data, err := resources.NewData(c.Period, fetchersRegistry)
	if err != nil {
		cancel()
		return nil, err
	}

	eval, err := evaluator.NewOpaEvaluator(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	// namespace will be passed as param from fleet on https://github.com/elastic/security-team/issues/2383 and it's user configurable
	resultsIndex := config.Datastream("", config.ResultsDatastreamIndexPrefix)
	if err != nil {
		cancel()
		return nil, err
	}

	t := transformer.NewTransformer(ctx, eval, resultsIndex)

	bt := &cloudbeat{
		ctx:         ctx,
		cancel:      cancel,
		config:      c,
		evaluator:   eval,
		data:        data,
		transformer: t,
	}
	return bt, nil
}

// Run starts cloudbeat.
func (bt *cloudbeat) Run(b *beat.Beat) error {
	logp.Info("cloudbeat is running! Hit CTRL-C to stop it.")

	if err := bt.data.Run(bt.ctx); err != nil {
		return err
	}

	procs, err := bt.configureProcessors(bt.config.Processors)
	if err != nil {
		return err
	}

	// Connect publisher (with beat's processors)
	if bt.client, err = b.Publisher.ConnectWith(beat.ClientConfig{
		Processing: beat.ProcessingConfig{
			Processor: procs,
		},
	}); err != nil {
		return err
	}

	output := bt.data.Output()

	for {
		select {
		case <-bt.ctx.Done():
			return nil
		case fetchedResources := <-output:
			cycleId, _ := uuid.NewV4()
			// update hidden-index that the beat's cycle has started
			bt.updateCycleStatus(cycleId, cycleStatusStart)
			cycleMetadata := transformer.CycleMetadata{CycleId: cycleId}
			// TODO: send events through a channel and publish them by a configured threshold & time
			events := bt.transformer.ProcessAggregatedResources(fetchedResources, cycleMetadata)
			bt.client.PublishAll(events)
			// update hidden-index that the beat's cycle has ended
			bt.updateCycleStatus(cycleId, cycleStatusEnd)
		}
	}
}

func InitRegistry(ctx context.Context, c config.Config) (resources.FetchersRegistry, error) {
	registry := resources.NewFetcherRegistry()
	err := resources.Factories.RegisterFetchers(registry, c)
	if err != nil {
		return nil, err
	}
	return registry, nil
}

// Stop stops cloudbeat.
func (bt *cloudbeat) Stop() {
	bt.data.Stop(bt.ctx, bt.cancel)
	bt.evaluator.Stop(bt.ctx)

	bt.client.Close()
}

// updateCycleStatus updates beat status in metadata ES index.
func (bt *cloudbeat) updateCycleStatus(cycleId uuid.UUID, status string) {
	metadataIndex := config.Datastream("", config.MetadataDatastreamIndexPrefix)
	cycleEndedEvent := beat.Event{
		Timestamp: time.Now(),
		Meta:      common.MapStr{libevents.FieldMetaIndex: metadataIndex},
		Fields: common.MapStr{
			"cycle_id": cycleId,
			"status":   status,
		},
	}
	bt.client.Publish(cycleEndedEvent)
}

// configureProcessors configure processors to be used by the beat
func (bt *cloudbeat) configureProcessors(processorsList processors.PluginConfig) (procs *processors.Processors, err error) {
	return processors.New(processorsList)
}
