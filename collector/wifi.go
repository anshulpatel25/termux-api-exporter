package collector

import (
	"context"
	"encoding/json"
	"log"

	"github.com/anshulpatel25/termux-api-exporter/executor"
	"github.com/anshulpatel25/termux-api-exporter/model"
	"github.com/prometheus/client_golang/prometheus"
)

// WiFiCollector collects WiFi metrics from termux-wifi-connectioninfo command
type WiFiCollector struct {
	executor      executor.Executor
	rssi          *prometheus.Desc
	linkSpeedMbps *prometheus.Desc
}

// NewWiFiCollector creates a new WiFiCollector
func NewWiFiCollector(exec executor.Executor) *WiFiCollector {
	return &WiFiCollector{
		executor: exec,
		rssi: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "wifi", "rssi_dbm"),
			"WiFi signal strength in dBm (RSSI)",
			nil,
			nil,
		),
		linkSpeedMbps: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "wifi", "link_speed_mbps"),
			"WiFi link speed in Mbps",
			nil,
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *WiFiCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rssi
	ch <- c.linkSpeedMbps
}

// Collect implements the prometheus.Collector interface
func (c *WiFiCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Execute the termux-wifi-connectioninfo command
	output, err := c.executor.Execute(ctx, "termux-wifi-connectioninfo")
	if err != nil {
		log.Printf("Error executing termux-wifi-connectioninfo: %v", err)
		return
	}

	// Parse the JSON output
	var info model.WiFiConnectionInfo
	if err := json.Unmarshal(output, &info); err != nil {
		log.Printf("Error parsing WiFi connection info JSON: %v", err)
		return
	}

	// Send metrics to Prometheus
	ch <- prometheus.MustNewConstMetric(
		c.rssi,
		prometheus.GaugeValue,
		float64(info.RSSI),
	)

	ch <- prometheus.MustNewConstMetric(
		c.linkSpeedMbps,
		prometheus.GaugeValue,
		float64(info.LinkSpeedMbps),
	)
}
