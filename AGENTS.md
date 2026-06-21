# Agent Development Guide

This document provides comprehensive guidance for AI agents working on the Termux API Exporter project.

## Project Overview

**Purpose**: A Prometheus exporter that collects metrics from Termux API commands on Android devices and exposes them in Prometheus format.

**Language**: Go 1.21+
**Architecture**: Clean Architecture with interface-based design
**Primary Dependencies**: `github.com/prometheus/client_golang`

## Core Architecture

### Layer Separation

The codebase follows clean architecture with three primary layers:

1. **Domain Layer** (`model/`)
   - Pure domain models with no external dependencies
   - Structs map to JSON output from Termux commands
   - Tags: Use `json:"field_name"` for JSON unmarshaling

2. **Business Logic Layer** (`collector/`, `executor/`)
   - Interfaces define contracts
   - Implementations provide concrete behavior
   - No direct dependencies between implementations

3. **Application Layer** (`main.go`)
   - Wires up dependencies
   - Handles HTTP server lifecycle
   - Registers collectors with Prometheus registry

### Key Design Patterns

- **Dependency Injection**: All components receive dependencies via constructors
- **Interface Segregation**: Small, focused interfaces (Executor, Collector)
- **Single Responsibility**: Each package has one clear purpose
- **Error Handling**: Always log errors but don't crash; return early on failures

## File Organization

```
.
├── model/              # Domain models (pure Go structs)
├── executor/           # Command execution abstraction
├── collector/          # Prometheus metric collectors
└── main.go            # Application entry point
```

### Naming Conventions

- **Packages**: Lowercase, singular nouns (`model`, `collector`, `executor`)
- **Files**: Lowercase with underscores if needed (`battery.go`, `wifi.go`)
- **Interfaces**: Describe capability (`Executor`, `Collector`)
- **Implementations**: Describe what/how (`TermuxExecutor`, `BatteryCollector`)
- **Constructors**: Always use `New` prefix (`NewBatteryCollector`)

## Common Development Tasks

### Adding a New Termux Command Collector

Follow this exact pattern for consistency:

#### Step 1: Create the Domain Model

**Location**: `model/<feature>.go`

```go
package model

// <Feature>Info represents the output from termux-<feature> command
type <Feature>Info struct {
    // Add fields that match JSON output from the command
    // Use appropriate Go types (int, float64, string, bool)
    FieldName Type `json:"json_field_name"`
}
```

**Example**: For `termux-battery-status`:
```go
type BatteryStatus struct {
	Present       bool    `json:"present"`
	Technology    string  `json:"technology"`
	Health        string  `json:"health"`
	Plugged       string  `json:"plugged"`
	Status        string  `json:"status"`
	Temperature   float64 `json:"temperature"`
	Voltage       int     `json:"voltage"`
	Current       int     `json:"current"`
	Percentage    int     `json:"percentage"`
	Level         int     `json:"level"`
	Scale         int     `json:"scale"`
	ChargeCounter int     `json:"charge_counter"`
	Cycle         int     `json:"cycle"`
}
```

#### Step 2: Create the Collector

**Location**: `collector/<feature>.go`

```go
package collector

import (
    "context"
    "encoding/json"
    "log"

    "github.com/anshulpatel25/termux-api-exporter/executor"
    "github.com/anshulpatel25/termux-api-exporter/model"
    "github.com/prometheus/client_golang/prometheus"
)

// <Feature>Collector collects metrics from termux-<feature> command
type <Feature>Collector struct {
    executor executor.Executor
    metric1  *prometheus.Desc
    metric2  *prometheus.Desc
}

// New<Feature>Collector creates a new <Feature>Collector
func New<Feature>Collector(exec executor.Executor) *<Feature>Collector {
    return &<Feature>Collector{
        executor: exec,
        metric1: prometheus.NewDesc(
            prometheus.BuildFQName(namespace, "<subsystem>", "<metric_name>"),
            "Metric description",
            nil,
            nil,
        ),
    }
}

// Describe implements the prometheus.Collector interface
func (c *<Feature>Collector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.metric1
    ch <- c.metric2
}

// Collect implements the prometheus.Collector interface
func (c *<Feature>Collector) Collect(ch chan<- prometheus.Metric) {
    ctx := context.Background()

    // Execute the termux command
    output, err := c.executor.Execute(ctx, "termux-<feature>")
    if err != nil {
        log.Printf("Error executing termux-<feature>: %v", err)
        return
    }

    // Parse JSON output
    var info model.<Feature>Info
    if err := json.Unmarshal(output, &info); err != nil {
        log.Printf("Error parsing <feature> JSON: %v", err)
        return
    }

    // Send metrics to Prometheus
    ch <- prometheus.MustNewConstMetric(
        c.metric1,
        prometheus.GaugeValue,
        float64(info.FieldName),
    )
}
```

#### Step 3: Register in main.go

Add these lines in `main.go`:

```go
// Create the collector (after other collectors)
<feature>Collector := collector.New<Feature>Collector(exec)

// Register the collector (after other registrations)
if err := registry.Register(<feature>Collector); err != nil {
    log.Fatalf("Failed to register <feature> collector: %v", err)
}
```

### Metric Naming Conventions

Follow Prometheus naming best practices:

- **Namespace**: Always `termux`
- **Subsystem**: Feature name (e.g., `battery`, `wifi`, `location`)
- **Name**: Descriptive metric name with unit suffix
- **Units**: Include in metric name (`_celsius`, `_mbps`, `_dbm`, `_bytes`)

**Examples**:
- `termux_battery_temperature_celsius`
- `termux_wifi_rssi_dbm`
- `termux_memory_available_bytes`

### Error Handling Guidelines

**Always follow this pattern**:

1. **Log errors, don't panic**: Use `log.Printf()` for errors
2. **Return early**: Exit the function on error, don't continue
3. **Add context**: Include command name or operation in error message
4. **Preserve errors**: Wrap errors with `fmt.Errorf("...: %w", err)`

```go
// Good ✅
output, err := c.executor.Execute(ctx, "termux-battery-status")
if err != nil {
    log.Printf("Error executing termux-battery-status: %v", err)
    return
}

// Bad ❌ - Don't panic
if err != nil {
    panic(err)
}
```

## Code Quality Standards

### Idiomatic Go

- Use pointer receivers for structs with methods
- Return errors, don't panic (except in `main.go` for fatal setup errors)
- Use `context.Background()` for top-level operations
- Short variable names in limited scope (`ch`, `err`, `ctx`)
- Descriptive names for package-level variables

### Comments

- Add doc comments for all exported types, functions, and methods
- Format: `// TypeName description` (no blank line before declaration)
- Explain "why" not "what" for complex logic
- Use complete sentences with proper punctuation

### Imports

Group imports in this order:
1. Standard library
2. External packages
3. Internal packages (this project)

```go
import (
    "context"
    "encoding/json"
    "log"

    "github.com/prometheus/client_golang/prometheus"

    "github.com/anshulpatel25/termux-api-exporter/executor"
    "github.com/anshulpatel25/termux-api-exporter/model"
)
```

## Testing Strategy

### Unit Testing with Mocks

Create mock implementations for testing:

```go
// executor/mock.go
package executor

import "context"

type MockExecutor struct {
    MockOutput []byte
    MockError  error
}

func (m *MockExecutor) Execute(ctx context.Context, command string, args ...string) ([]byte, error) {
    return m.MockOutput, m.MockError
}
```

### Test File Naming

- Test files: `<filename>_test.go`
- Place in same package as code under test
- Use table-driven tests for multiple scenarios

## Common Pitfalls to Avoid

### 1. Don't Mix Concerns

❌ **Bad**: Putting HTTP handling in collector
```go
func (c *BatteryCollector) Collect(w http.ResponseWriter, r *http.Request) {
    // NO!
}
```

✅ **Good**: Keep collectors focused on metrics
```go
func (c *BatteryCollector) Collect(ch chan<- prometheus.Metric) {
    // YES!
}
```

### 2. Don't Ignore Errors

❌ **Bad**: Silent failures
```go
output, _ := c.executor.Execute(ctx, "termux-battery-status")
```

✅ **Good**: Log and return
```go
output, err := c.executor.Execute(ctx, "termux-battery-status")
if err != nil {
    log.Printf("Error: %v", err)
    return
}
```

### 3. Don't Hardcode Values

❌ **Bad**: Magic numbers and strings
```go
time.Sleep(5 * time.Second)
```

✅ **Good**: Use constants or configuration
```go
const defaultTimeout = 5 * time.Second
```

### 4. Don't Break Interface Contracts

All collectors must implement both methods:
- `Describe(chan<- *prometheus.Desc)`
- `Collect(chan<- prometheus.Metric)`

### 5. Don't Use panic in Business Logic

❌ **Bad**: Panicking on errors
```go
if err != nil {
    panic(err)
}
```

✅ **Good**: Log and continue or return
```go
if err != nil {
    log.Printf("Error: %v", err)
    return
}
```

## Termux Command Reference

### Available Commands

Current collectors use:
- `termux-battery-status` - Battery metrics
- `termux-wifi-connectioninfo` - WiFi metrics

### Common Termux Commands for Future Collectors

- `termux-location` - GPS location (lat/long/altitude)
- `termux-sensor` - Device sensors (accelerometer, gyroscope, etc.)
- `termux-telephony-deviceinfo` - Phone/SIM info
- `termux-clipboard-get` - Clipboard contents
- `termux-torch` - Flashlight control
- `termux-brightness` - Screen brightness

### Testing Commands Locally

Use the mock scripts in the repository root:
- `termux-battery-status` - Returns mock battery JSON
- `termux-wifi-connectioninfo` - Returns mock WiFi JSON

## Build and Run

### Build Commands

```bash
# Standard build
go build -o termux-api-exporter

# With optimizations
go build -ldflags="-s -w" -o termux-api-exporter

# Cross-compile for Android ARM64
GOOS=linux GOARCH=arm64 go build -o termux-api-exporter
```

### Development Workflow

1. Make changes to code
2. Run `go mod tidy` if adding/removing dependencies
3. Build: `go build -o termux-api-exporter`
4. Test: `./termux-api-exporter`
5. Verify metrics: `curl http://localhost:9797/metrics`

## Prometheus Integration

### Metric Types

Use appropriate Prometheus metric types:

- **Gauge**: Values that go up and down (temperature, percentage, RSSI)
- **Counter**: Monotonically increasing values (total requests, errors)
- **Histogram**: Observations (request durations, response sizes)
- **Summary**: Similar to histogram with quantiles

For Termux metrics, **Gauge is most common** since we're exposing current state.

### Metric Registration

Always check registration errors:

```go
if err := registry.Register(collector); err != nil {
    log.Fatalf("Failed to register collector: %v", err)
}
```

## Documentation Updates

When adding new features, update:

1. **README.md**:
   - Features list
   - Architecture diagram
   - Metrics table
   - Example output

2. **This file (AGENTS.md)**:
   - Add to "Available Commands" if new Termux command
   - Update examples if patterns change

3. **Comments in code**:
   - Doc comments for new types
   - Inline comments for complex logic

## Quick Reference

### File Templates

See "Adding a New Termux Command Collector" section above for complete templates.

### Key Constants

```go
const (
    namespace = "termux"           // Prometheus namespace
    defaultTimeout = 5 * time.Second
    defaultPort = ":9797"
    metricsPath = "/metrics"
)
```

### Essential Imports

```go
// For models
import "encoding/json"

// For collectors
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/anshulpatel25/termux-api-exporter/executor"
    "github.com/anshulpatel25/termux-api-exporter/model"
)

// For main
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
)
```

## Getting Help

- Check existing collectors (battery, wifi) as reference implementations
- Follow the exact patterns shown in this document
- Prometheus docs: https://prometheus.io/docs/
- Go best practices: https://go.dev/doc/effective_go

## Summary

**Core Principle**: Follow the existing patterns exactly. Every new collector should look structurally identical to existing ones, just with different:
- Model fields
- Termux command name
- Metric names and descriptions

This consistency makes the codebase easy to understand and maintain.
