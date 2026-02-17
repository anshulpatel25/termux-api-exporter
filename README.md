# Termux API Exporter

A Prometheus exporter for Termux API metrics, written in Go following clean architecture principles.

## Features

- ✅ Exposes battery temperature and percentage as Prometheus metrics
- ✅ Exposes WiFi signal strength (RSSI) and link speed metrics
- ✅ Clean architecture with clear separation of concerns
- ✅ Extensible design for adding more Termux API commands
- ✅ Proper error handling and logging
- ✅ Graceful shutdown support
- ✅ Health check endpoint
- ✅ Configurable timeout and server settings

## Architecture

The project follows clean architecture principles with clear separation:

```
├── main.go                 # Application entry point
├── model/
│   ├── battery.go         # Battery domain model
│   └── wifi.go            # WiFi domain model
├── executor/
│   ├── executor.go        # Command executor interface
│   └── termux.go          # Termux command implementation
└── collector/
    ├── collector.go       # Collector interface
    ├── battery.go         # Battery metrics collector
    └── wifi.go            # WiFi metrics collector
```

### Design Principles

- **Interface-based design**: Core components use interfaces for loose coupling
- **Dependency injection**: Dependencies are injected, making testing easier
- **Single Responsibility**: Each package has a clear, focused purpose
- **Extensibility**: Easy to add new collectors for other Termux commands

## Prerequisites

### System Requirements

**This project is designed for rooted Android devices.**

The device requires:
- **Rooted Android device** with Magisk or similar root solution
- **Termux** app installed from F-Droid (not Play Store)
- **Termux:API** app installed from F-Droid
- **Superuser permissions** granted to Termux and Termux:API via Magisk or similar
- Go 1.21 or higher (for building from source)

## Installation

### Step 1: Grant Superuser Permissions

Before installing the exporter, ensure that both Termux and Termux:API have superuser permissions:

1. Open **Magisk Manager** (or your root solution)
2. Navigate to the **Superuser** section
3. Grant root access to:
   - **Termux** app
   - **Termux:API** app

### Step 2: Install Required Termux Packages

Open Termux and install the required packages:

```bash
# Update package repositories
pkg update && pkg upgrade -y

# Install sudo package
pkg install sudo

# Install termux-api package
pkg install termux-api

# Install termux-services package
pkg install termux-services

# Restart Termux after installing termux-services
exit
# Then reopen Termux
```

### Step 3: Install the Exporter

Install the exporter in your home directory:

```bash
# Navigate to home directory
cd $HOME

# Clone the repository
git clone https://github.com/anshulpatel25/termux-api-exporter.git

# Navigate to the project directory
cd termux-api-exporter

# Download dependencies
go mod download

# Build the binary
export GOOS=linux
export GOARCH=arm64
go build -o termux-api-exporter
```

### Step 4: Create Service Configuration

Create a service file to run the exporter as a background service:

```bash
# Create the service directory
mkdir -p $PREFIX/var/service/termux-api-exporter

# Create the run script
cat > $PREFIX/var/service/termux-api-exporter/run << 'EOF'
#!/data/data/com.termux/files/usr/bin/sh
exec $HOME/termux-api-exporter/termux-api-exporter 2>&1
EOF

# Make the run script executable
chmod +x $PREFIX/var/service/termux-api-exporter/run
```

### Step 5: Enable and Start the Service

Enable the service to start automatically:

```bash
# Enable the service
sv-enable termux-api-exporter

# The service will start automatically
# Check service status
sv status termux-api-exporter

# To manually start/stop/restart the service:
# sv up termux-api-exporter     # Start
# sv down termux-api-exporter   # Stop
# sv restart termux-api-exporter # Restart
```

### Verification

Verify the exporter is running:

```bash
# Check if the service is running
sv status termux-api-exporter

# Test the metrics endpoint
curl http://localhost:9797/metrics

# Test the health endpoint
curl http://localhost:9797/health
```

## Usage

### Running the Exporter

```bash
# Run with default settings (port 9100)
./termux-api-exporter

# Run with custom port
./termux-api-exporter -listen-address=":9797"

# Run with custom metrics path
./termux-api-exporter -metrics-path="/custom-metrics"

# Run with custom timeout
./termux-api-exporter -timeout=10s
```

### Command-line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-listen-address` | `:9797` | Address to listen on for HTTP requests |
| `-metrics-path` | `/metrics` | Path under which to expose metrics |
| `-timeout` | `5s` | Timeout for executing termux commands |

### Endpoints

- `/` - Landing page with links
- `/metrics` - Prometheus metrics endpoint
- `/health` - Health check endpoint

## Metrics

The exporter exposes the following metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `termux_battery_temperature_celsius` | Gauge | Battery temperature in Celsius |
| `termux_battery_percentage` | Gauge | Battery charge percentage (0-100) |
| `termux_wifi_rssi_dbm` | Gauge | WiFi signal strength in dBm (RSSI) |
| `termux_wifi_link_speed_mbps` | Gauge | WiFi link speed in Mbps |

### Example Output

```
# HELP termux_battery_percentage Battery charge percentage
# TYPE termux_battery_percentage gauge
termux_battery_percentage 65

# HELP termux_battery_temperature_celsius Battery temperature in Celsius
# TYPE termux_battery_temperature_celsius gauge
termux_battery_temperature_celsius 28.6

# HELP termux_wifi_link_speed_mbps WiFi link speed in Mbps
# TYPE termux_wifi_link_speed_mbps gauge
termux_wifi_link_speed_mbps 173

# HELP termux_wifi_rssi_dbm WiFi signal strength in dBm (RSSI)
# TYPE termux_wifi_rssi_dbm gauge
termux_wifi_rssi_dbm -30
```

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'termux-api'
    static_configs:
      - targets: ['localhost:9100']
```

## Extending for More Termux Commands

The architecture makes it easy to add support for additional Termux commands:

### 1. Add a new model (if needed)

```go
// model/location.go
package model

type Location struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Altitude  float64 `json:"altitude"`
}
```

### 2. Create a new collector

```go
// collector/location.go
package collector

import (
    "context"
    "encoding/json"
    "log"

    "github.com/anshulpatel25/termux-api-exporter/executor"
    "github.com/anshulpatel25/termux-api-exporter/model"
    "github.com/prometheus/client_golang/prometheus"
)

type LocationCollector struct {
    executor  executor.Executor
    latitude  *prometheus.Desc
    longitude *prometheus.Desc
}

func NewLocationCollector(exec executor.Executor) *LocationCollector {
    return &LocationCollector{
        executor: exec,
        latitude: prometheus.NewDesc(
            prometheus.BuildFQName("termux", "location", "latitude"),
            "Current latitude",
            nil, nil,
        ),
        longitude: prometheus.NewDesc(
            prometheus.BuildFQName("termux", "location", "longitude"),
            "Current longitude",
            nil, nil,
        ),
    }
}

func (c *LocationCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.latitude
    ch <- c.longitude
}

func (c *LocationCollector) Collect(ch chan<- prometheus.Metric) {
    ctx := context.Background()
    output, err := c.executor.Execute(ctx, "termux-location")
    if err != nil {
        log.Printf("Error executing termux-location: %v", err)
        return
    }

    var location model.Location
    if err := json.Unmarshal(output, &location); err != nil {
        log.Printf("Error parsing location JSON: %v", err)
        return
    }

    ch <- prometheus.MustNewConstMetric(c.latitude, prometheus.GaugeValue, location.Latitude)
    ch <- prometheus.MustNewConstMetric(c.longitude, prometheus.GaugeValue, location.Longitude)
}
```

### 3. Register the collector in main.go

```go
// Add to main.go
locationCollector := collector.NewLocationCollector(exec)
if err := registry.Register(locationCollector); err != nil {
    log.Fatalf("Failed to register location collector: %v", err)
}
```

## Error Handling

The exporter includes comprehensive error handling:

- **Command execution errors**: Logged with details about exit codes and stderr
- **Timeout errors**: Commands that exceed the timeout are terminated
- **JSON parsing errors**: Invalid JSON output is logged without crashing
- **Graceful shutdown**: SIGTERM/SIGINT signals are handled gracefully

## Development

### Project Structure

```
.
├── collector/              # Metric collectors
│   ├── collector.go       # Collector interface
│   └── battery.go         # Battery collector implementation
├── executor/              # Command execution layer
│   ├── executor.go        # Executor interface
│   └── termux.go          # Termux executor implementation
├── model/                 # Domain models
│   └── battery.go         # Battery status model
├── main.go                # Application entry point
├── go.mod                 # Go module file
└── README.md              # This file
```

### Testing

You can test the exporter without Termux by creating a mock executor:

```go
// executor/mock.go
package executor

import (
    "context"
)

type MockExecutor struct {
    output []byte
    err    error
}

func NewMockExecutor(output []byte, err error) *MockExecutor {
    return &MockExecutor{output: output, err: err}
}

func (m *MockExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
    return m.output, m.err
}
```

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
