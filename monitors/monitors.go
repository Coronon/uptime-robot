package monitors

import (
	"fmt"
	"net/http"

	"github.com/Coronon/uptime-robot/config"
	"go.uber.org/zap"
)

type Monitor interface {
	// Name of this monitor (user defined in config)
	Name() string
	// Type of this monitor
	Type() string
	// Resolved host of this monitor
	HostURL() string
	// Interval in seconds this monitor runs
	Interval() int

	// Start running this monitor
	Run()
}

// Schedule monitors to run in background
func SetupMonitors(c config.Config) {
	// Check at least one monitor defined
	numMonitors := len(c.Monitors)
	if numMonitors == 0 {
		zap.L().Panic("No monitors defined")
	}

	zap.S().Infow("Setting up monitors",
		"count", len(c.Monitors),
	)

	// Actually setup monitors based on config
	monitors := make([]Monitor, len(c.Monitors))

	for i := range c.Monitors {
		monitor := &c.Monitors[i]

		zap.S().Debugw("Setting up monitor",
			"name", monitor.Name,
			"type", monitor.Type,
		)

		// Determine host
		var hostURL string
		for h := range c.Hosts {
			host := &c.Hosts[h]

			if host.Name == monitor.Host {
				zap.S().Debugw("Found matching host",
					"name", host.Name,
					"url", host.URL,
				)
				hostURL = host.URL
				break
			}
		}
		if hostURL == "" {
			zap.S().Panicw("Could not find host",
				"host", monitor.Host,
			)
		}

		// Ensure host ends with a trailing '/'
		if hostURL[len(hostURL)-1:] != "/" {
			zap.S().Debugw("Adding trailing '/' to host url",
				"host", monitor.Host,
				"old_url", hostURL,
			)
			hostURL = hostURL + "/"
		}

		// Setup based on monitor type
		switch monitor.Type {
		case "alive":
			monitors[i] = setupAliveMonitor(hostURL, monitor)
		default:
			zap.S().Panicw("Unknown monitor type",
				"type", monitor.Type,
			)
		}
	}

	// Run monitors
	zap.L().Info("Starting monitors...")
	for i := range monitors {
		monitors[i].Run()
	}
	zap.L().Info("All monitors started")
}

// Represents an up/down monitor monitorStatus
type monitorStatus string

const (
	StatusUp   monitorStatus = "up"
	StatusDown monitorStatus = "down"
)

// Pushes a monitors state to an uptime host handling creation of the correctly
// formatted URL
//
// Returns the HTTP requests response/error
func pushToHost(
	host string,
	key string,
	status monitorStatus,
	message string,
	pingMs int,
) (resp *http.Response, err error) {
	//? We already ensure that host ends with a trailing / in SetupMonitors
	url := fmt.Sprintf("%v%v?status=%v&msg=%v&ping=%v", host, key, status, message, pingMs)

	zap.S().Debugw("Pushing to host",
		"host", host,
		"key", key,
		"status", status,
		"message", message,
		"pingMs", pingMs,
		"url", url,
	)

	return http.Get(url)
}
