package main

import (
	"github.com/kardianos/service"
	"go.uber.org/zap"

	"github.com/Coronon/uptime-robot/config"
	"github.com/Coronon/uptime-robot/monitors"
)

const version = "v1.0.0"
const serviceName = "de.rubinraithel.uptime-robot"
const displayName = "Uptime-Robot" + " " + version
const serviceDesc = "Utility service that provides push based uptime monitoring for various services"

type program struct{}

func (p program) Start(s service.Service) error {
	zap.S().Info(s.String() + " started")
	go p.run()
	return nil
}

func (p program) Stop(s service.Service) error {
	zap.S().Info(s.String() + " stopped")
	return nil
}

func (p program) run() {
	// Handle config
	// TODO: Make this dynamic with a default
	configPath := "example.config.yml"

	zap.S().Info("Parsing config at: %v", configPath)
	config := config.ReadConfig(configPath)
	zap.S().Info("Got assigned node name: %v", config.NodeName)

	// Setup monitors
	monitors.SetupMonitors(config)
}

func init() {
	// Setup logging
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)
}

func main() {
	// Setup service
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: displayName,
		Description: serviceDesc,
	}

	prg := &program{}
	s, err := service.New(prg, serviceConfig)

	if err != nil {
		zap.S().Fatal("Cannot create the service", err.Error())
	}
	err = s.Run()
	if err != nil {
		zap.S().Fatal("Cannot start the service", err.Error())
	}
}
