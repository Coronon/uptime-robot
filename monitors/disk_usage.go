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

	// Linux: stat -f {fileInFilesystem}, Windows: drive letter
	filePath string
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
		"file_path", m.filePath,
		"down_threshold", m.downThreshold,
	)

	diskInfo := du.NewDiskUsage(m.filePath)

	if diskInfo == nil || math.IsNaN(float64(diskInfo.Usage())) {
		zap.S().Errorw("Error getting disk usage",
			"name", m.name,
			"disk_space", diskInfo,
			// This usage may be inaccurate, see below
			"disk_usage", diskInfo.Usage(),
		)

		// We want to still push this error to the uptime host
		return StatusDown, "Error getting disk usage", 0, nil
	}

	// The normal .Usage() uses the complete .Free() instead of the actually
	// usable .Available() space, so we compute what df would output here
	// Basically stat -> Blocks Free vs Blocks Available
	percentageAvailable := float64(diskInfo.Available()) / float64(diskInfo.Size()) * 100
	// This is a primitive round as we know that usage can never be negative
	percentage := 100 - int(percentageAvailable+0.5)

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
	if monitor.FilePath == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "file_path",
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
		filePath:      monitor.FilePath,
		downThreshold: monitor.DownThreshold,
	}
}
