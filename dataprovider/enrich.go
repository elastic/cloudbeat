package dataprovider

import "github.com/elastic/beats/v7/libbeat/beat"

type ElasticCommonDataProvider interface {
	GetElasticCommonData() (map[string]interface{}, error)
}

type enricher struct {
	dataprovider ElasticCommonDataProvider
}

func NewEnricher(dataprovider ElasticCommonDataProvider) *enricher {
	return &enricher{
		dataprovider: dataprovider,
	}
}

func (e *enricher) EnrichEvent(event *beat.Event) error {
	ecsData, err := e.dataprovider.GetElasticCommonData()
	if err != nil {
		return err
	}

	for k, v := range ecsData {
		_, err := event.Fields.Put(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
