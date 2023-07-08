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
	Name                 string `yaml:"name"`
	Type                 string `yaml:"type"`
	Host                 string `yaml:"host"`
	Key                  string `yaml:"key"`
	Interval             int    `yaml:"interval"`
	Timeout              int    `yaml:"timeout,omitempty"`
	FilePath             string `yaml:"file_path,omitempty"`
	DownThreshold        int    `yaml:"down_threshold,omitempty"`
	SMTPHost             string `yaml:"smtp_host,omitempty"`
	SMTPPort             int    `yaml:"smtp_port,omitempty"`
	SMTPSenderAddress    string `yaml:"smtp_sender_address,omitempty"`
	SMTPRecipientAddress string `yaml:"smtp_recipient_address,omitempty"`
	SMTPUsername         string `yaml:"smtp_username,omitempty"`
	SMTPPassword         string `yaml:"smtp_password,omitempty"`
	IMAPHost             string `yaml:"imap_host,omitempty"`
	IMAPPort             int    `yaml:"imap_port,omitempty"`
	IMAPUsername         string `yaml:"imap_username,omitempty"`
	IMAPPassword         string `yaml:"imap_password,omitempty"`
	MessageSubject       string `yaml:"message_subject,omitempty"`
	MessageBody          string `yaml:"message_body,omitempty"`
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
