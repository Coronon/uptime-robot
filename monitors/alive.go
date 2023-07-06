package monitors

import (
	"github.com/coronon/uptime-robot/config"
)

type aliveMonitor struct {
	name     string
	host     string
	key      string
	interval int
}

func (m *aliveMonitor) Name() string {
	return m.name
}

func (m *aliveMonitor) Type() string {
	return "alive"
}

func (m *aliveMonitor) HostURL() string {
	return m.host
}

func (m *aliveMonitor) Key() string {
	return m.key
}

func (m *aliveMonitor) Interval() int {
	return m.interval
}

func (m *aliveMonitor) Run() (monitorStatus, string, int, error) {
	// Simply let the upstream host know that we are alive
	return StatusUp, "OK", 0, nil
}

// Setup a monitor of type 'alive'
func setupAliveMonitor(host string, monitor *config.Monitor) *aliveMonitor {
	return &aliveMonitor{name: monitor.Name, host: host, interval: monitor.Interval, key: monitor.Key}
}
