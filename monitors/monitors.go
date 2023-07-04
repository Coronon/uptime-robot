package monitors

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

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
	// Key used to identify this monitor on the uptime host
	Key() string
	// Interval in seconds this monitor runs
	Interval() int

	// Run a single iteration of this monitor (periodically called)
	Run() (status monitorStatus, message string, pingMs int, err error)
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
	monitorKeys := make([]string, len(c.Monitors))

	for i := range c.Monitors {
		monitor := &c.Monitors[i]

		zap.S().Debugw("Setting up monitor",
			"name", monitor.Name,
			"type", monitor.Type,
		)

		// Check key not reused
		for k := 0; k < i; k++ {
			if monitorKeys[k] == monitor.Key {
				zap.S().Panicw("Key is not unique",
					"monitor", monitor.Name,
					"key", monitor.Key,
				)
			}
		}
		monitorKeys[i] = monitor.Key
		zap.S().Debugw("Key is unique",
			"monitor", monitor.Name,
			"key", monitor.Key,
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
		case "disk_usage":
			monitors[i] = setupdiskUsageMonitor(hostURL, monitor)
		default:
			zap.S().Panicw("Unknown monitor type",
				"type", monitor.Type,
			)
		}
	}

	// Run monitors
	zap.L().Info("Starting monitors...")
	for i := range monitors {
		runMonitorPeriodically(monitors[i])
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
	// Parse base URL
	baseUrl, err := url.Parse(host)
	if err != nil {
		zap.S().DPanic("Malformed host URL",
			"url", host,
			"error", err.Error(),
		)
		return nil, err
	}

	// Add key path
	//? We already ensure that host ends with a trailing / in SetupMonitors
	baseUrl.Path += key

	// Add dynamic status information
	params := url.Values{}
	params.Add("status", string(status))
	params.Add("msg", message)
	params.Add("ping", fmt.Sprint(pingMs))
	baseUrl.RawQuery = params.Encode()

	// Build final URL
	url := baseUrl.String()

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

// Run a monitor periodically based on its configured interval
//
// Should be called in a go-routine
func runMonitorPeriodically(m Monitor) {
	sleepTime := time.Duration(m.Interval()) * time.Second

	for {
		go func() {
			zap.S().Debugw("Running monitor",
				"name", m.Name(),
				"type", m.Type(),
				"host", m.HostURL(),
				"key", m.Key(),
				"interval", m.Interval(),
			)

			status, message, ping, err := m.Run()
			if err != nil {
				zap.S().Warnw("Error running monitor",
					"name", m.Name(),
					"type", m.Type(),
					"host", m.HostURL(),
					"key", m.Key(),
					"interval", m.Interval(),
					"error", err,
				)
			} else {
				// Only push to host if monitor did not error (down should not be an error)
				resp, err := pushToHost(m.HostURL(), m.Key(), status, message, ping)

				if err != nil {
					zap.S().Warnw("Error pushing to host",
						"name", m.Name(),
						"type", m.Type(),
						"host", m.HostURL(),
						"key", m.Key(),
						"interval", m.Interval(),
						"error", err,
					)
				} else if resp.StatusCode != 200 {
					zap.S().Warnw("Error pushing to host",
						"name", m.Name(),
						"type", m.Type(),
						"host", m.HostURL(),
						"key", m.Key(),
						"interval", m.Interval(),
						"resp_statuscode", resp.StatusCode,
					)
				}
			}
		}()

		// Always wait for interval
		//? The interval does not depend on the time the monitor and pushing it's
		//? result take
		time.Sleep(sleepTime)
	}
}
