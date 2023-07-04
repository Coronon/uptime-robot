package monitors

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Coronon/uptime-robot/config"
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
		log.Fatal("No monitors defined")
	}

	log.Printf("Setting up %v monitors", len(c.Monitors))

	// Actually setup monitors based on config
	monitors := make([]Monitor, len(c.Monitors))

	for i := range c.Monitors {
		monitor := &c.Monitors[i]

		log.Printf("Setting up monitor: %v (%v)", monitor.Name, monitor.Type)

		// Determine host
		var hostURL string
		for h := range c.Hosts {
			host := &c.Hosts[h]

			if host.Name == monitor.Host {
				hostURL = host.URL
				break
			}
		}
		if hostURL == "" {
			log.Fatalf("Could not find host: %v", monitor.Host)
		}

		// Ensure host ends with a trailing '/'
		if hostURL[len(hostURL)-1:] != "/" {
			hostURL = hostURL + "/"
		}

		// Setup based on monitor type
		switch monitor.Type {
		case "alive":
			monitors[i] = setupAliveMonitor(hostURL, monitor)
		default:
			log.Fatalf("Unknown monitor type: %v", monitor.Type)
		}
	}

	// Run monitors
	log.Print("Starting monitors...")
	for i := range monitors {
		monitors[i].Run()
	}
	log.Print("All monitors started")
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

	return http.Get(url)
}
