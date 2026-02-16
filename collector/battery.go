package collector

import (
	"context"
	"encoding/json"
	"log"

	"github.com/anshulpatel25/termux-api-exporter/executor"
	"github.com/anshulpatel25/termux-api-exporter/model"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "termux"
	subsystem = "battery"
)

// BatteryCollector collects battery metrics from termux-battery-status command
type BatteryCollector struct {
	executor    executor.Executor
	temperature *prometheus.Desc
	percentage  *prometheus.Desc
}

// NewBatteryCollector creates a new BatteryCollector
func NewBatteryCollector(exec executor.Executor) *BatteryCollector {
	return &BatteryCollector{
		executor: exec,
		temperature: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "temperature_celsius"),
			"Battery temperature in Celsius",
			nil,
			nil,
		),
		percentage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "percentage"),
			"Battery charge percentage",
			nil,
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *BatteryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.temperature
	ch <- c.percentage
}

// Collect implements the prometheus.Collector interface
func (c *BatteryCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Execute the termux-battery-status command
	output, err := c.executor.Execute(ctx, "termux-battery-status")
	if err != nil {
		log.Printf("Error executing termux-battery-status: %v", err)
		return
	}

	// Parse the JSON output
	var status model.BatteryStatus
	if err := json.Unmarshal(output, &status); err != nil {
		log.Printf("Error parsing battery status JSON: %v", err)
		return
	}

	// Send metrics to Prometheus
	ch <- prometheus.MustNewConstMetric(
		c.temperature,
		prometheus.GaugeValue,
		status.Temperature,
	)

	ch <- prometheus.MustNewConstMetric(
		c.percentage,
		prometheus.GaugeValue,
		float64(status.Percentage),
	)
}
