package monitors

import (
	"time"

	"github.com/Coronon/uptime-robot/config"
	"go.uber.org/zap"
)

type aliveMonitor struct {
	name     string
	host     string
	interval int
	key      string
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

func (m *aliveMonitor) Interval() int {
	return m.interval
}

func (m *aliveMonitor) Run() {
	go m.run()
}

func (m *aliveMonitor) run() {
	sleepTime := time.Duration(m.interval) * time.Second

	for {
		// Simply let the upstream host know that we are alive
		go func() {
			zap.S().Debugw("Running monitor",
				"name", m.name,
				"type", "alive",
				"host", m.host,
				"key", m.key,
			)

			_, err := pushToHost(m.host, m.key, StatusUp, "OK", 0)

			if err != nil {
				zap.S().Warnw("Error while pushing to host",
					"name", m.name,
					"type", "alive",
					"host", m.host,
					"key", m.key,
				)
			}
		}()

		// Wait for interval
		time.Sleep(sleepTime)
	}
}

// Setup a monitor of type 'alive'
func setupAliveMonitor(host string, monitor *config.Monitor) *aliveMonitor {
	return &aliveMonitor{name: monitor.Name, host: host, interval: monitor.Interval, key: monitor.Key}
}
