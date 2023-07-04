package monitors

import (
	"fmt"
	"math"

	"github.com/ricochet2200/go-disk-usage/du"
	"go.uber.org/zap"

	"github.com/Coronon/uptime-robot/config"
)

type diskUsageMonitor struct {
	name     string
	host     string
	key      string
	interval int

	// Linux: df {fileSystem}, Windows: drive letter
	fileSystem string
	// Percentage of used space which will start triggering down status
	downThreshold int
}

func (m *diskUsageMonitor) Name() string {
	return m.name
}

func (m *diskUsageMonitor) Type() string {
	return "disk_usage"
}

func (m *diskUsageMonitor) HostURL() string {
	return m.host
}

func (m *diskUsageMonitor) Key() string {
	return m.key
}

func (m *diskUsageMonitor) Interval() int {
	return m.interval
}

func (m *diskUsageMonitor) Run() (monitorStatus, string, int, error) {
	// Get disk usage
	zap.S().Debugw("Getting disk usage",
		"name", m.name,
		"file_system", m.fileSystem,
		"down_threshold", m.downThreshold,
	)

	usage := du.NewDiskUsage(m.fileSystem)

	if usage == nil || math.IsNaN(float64(usage.Usage())) {
		zap.S().Errorw("Error getting disk usage",
			"name", m.name,
			"disk_space", usage,
			"usage", usage.Usage(),
		)

		// We want to still push this error to the uptime host
		return StatusDown, "Error getting disk usage", 0, nil
	}

	// This is a primitive round as we know that usage can never be negative
	percentage := int((usage.Usage() * 100) + 0.5)

	var status monitorStatus
	var message string
	if percentage < m.downThreshold {
		status = StatusUp
		message = "OK"
	} else {
		status = StatusDown
		message = fmt.Sprintf("Exceeds threshold of %v%%", m.downThreshold)
	}

	zap.S().Debugw("Got disk usage",
		"name", m.name,
		"status", status,
		"percentage", percentage,
		"message", message,
	)

	return status, message, percentage, nil
}

// Setup a monitor of type 'alive'
func setupdiskUsageMonitor(host string, monitor *config.Monitor) *diskUsageMonitor {
	if monitor.FileSystem == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "file_system",
		)
	}

	if monitor.DownThreshold == 0 {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "down_threshold",
		)
	}

	return &diskUsageMonitor{
		name:          monitor.Name,
		host:          host,
		interval:      monitor.Interval,
		key:           monitor.Key,
		fileSystem:    monitor.FileSystem,
		downThreshold: monitor.DownThreshold,
	}
}
