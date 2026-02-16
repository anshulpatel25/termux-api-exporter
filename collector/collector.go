package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Collector defines the interface for metric collectors
type Collector interface {
	// Describe sends the super-set of all possible descriptors of metrics
	// collected by this Collector to the provided channel
	Describe(ch chan<- *prometheus.Desc)

	// Collect is called by the Prometheus registry when collecting metrics
	Collect(ch chan<- prometheus.Metric)
}
