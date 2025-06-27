package observability

import "github.com/prometheus/client_golang/prometheus"

const (
	namespace = "orestis-cloudbeat"
)

var (
	EventsPublished = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "events.published",
			Help:      "Number of events published successfully.",
		},
	)

	All = []prometheus.Collector{
		EventsPublished,
	}
)
