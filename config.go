package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NodeName string `yaml:"node_name"`
	Hosts    []struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	} `yaml:"hosts"`
	Monitors []struct {
		Name          string `yaml:"name"`
		Type          string `yaml:"type"`
		Interval      int    `yaml:"interval"`
		FileSystem    string `yaml:"file_system,omitempty"`
		DownThreshold int    `yaml:"down_threshold,omitempty"`
		Host          string `yaml:"host,omitempty"`
		Key           string `yaml:"key,omitempty"`
	} `yaml:"monitors"`
}

// Read and parse a yaml config at path
func ReadConfig(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	c := Config{}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	return c
}
