package config

import (
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Generated with:
// https://zhwt.github.io/yaml-to-go/
type Config struct {
	NodeName string    `yaml:"node_name"`
	Hosts    []Host    `yaml:"hosts"`
	Monitors []Monitor `yaml:"monitors"`
}
type Host struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}
type Monitor struct {
	Name          string `yaml:"name"`
	Type          string `yaml:"type"`
	Host          string `yaml:"host"`
	Key           string `yaml:"key"`
	Interval      int    `yaml:"interval"`
	FilePath      string `yaml:"file_path,omitempty"`
	DownThreshold int    `yaml:"down_threshold,omitempty"`
}

// Read and parse a yaml config at path
func ReadConfig(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		zap.S().Fatalf("Error reading config: %v", err)
	}

	c := Config{}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		zap.S().Fatalf("Error parsing config: %v", err)
	}

	return c
}
