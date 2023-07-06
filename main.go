package main

import (
	"flag"
	"os"

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
	configPath := "uptime-robot.yml"

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
	// Handle command line arguments
	shouldInstall := flag.Bool("install", false, "Installs Uptime-Robot as a service on your computer")
	shouldUninstall := flag.Bool("uninstall", false, "Uninstalls the Uptime-Robot service from your computer")
	isForcedRun := flag.Bool("interactive", false, "Run Uptime-Robot interactively (not as a service)")

	flag.Parse()

	// Setup service
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: displayName,
		Description: serviceDesc,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		zap.S().Fatal("Cannot create the service configuration", err.Error())
	}

	// Handle uninstall
	if *shouldUninstall {
		zap.L().Info("Uninstalling the Uptime-Robot service from your computer...")

		err = s.Uninstall()
		if err != nil {
			zap.S().Fatalw("Cannot uninstall the service", "error", err)
		}
		zap.L().Info("Uninstalled the service!")
		os.Exit(0)
	}

	// Handle install
	// This is handled after uninstall to allow "re-installing" by specifying both flags
	if *shouldInstall {
		zap.L().Info("Installing the Uptime-Robot service to your computer...")

		err = s.Install()
		if err != nil {
			zap.S().Fatalw("Cannot install the service", "error", err)
		}
		zap.L().Info("Installed the service!")
		os.Exit(0)
	}

	// Actually run program if invoked by service manager or forced interactively
	if !service.Interactive() || *isForcedRun {
		zap.L().Info("Starting Uptime-Robot...")
		err = s.Run()
		if err != nil {
			zap.S().Fatalw("Cannot start", "error", err.Error())
		}
	} else {
		zap.L().Info("Not running under a service manager or forced interactive - not starting Uptime-Robot")
	}
}
