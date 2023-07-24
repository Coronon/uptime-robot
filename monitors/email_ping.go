package monitors

import (
	"github.com/coronon/uptime-robot/config"
	"go.uber.org/zap"
)

type emailPingMonitor struct {
	name     string
	host     string
	key      string
	interval int

	smtp_host              string
	smtp_port              int
	smtp_force_tls         bool
	smtp_sender_address    string
	smtp_recipient_address string
	smtp_username          string
	smtp_password          string

	imap_host      string
	imap_port      int
	imap_force_tls bool
	imap_username  string
	imap_password  string

	message_subject  string
	message_body     string
	response_subject string

	timeout int
}

func (m *emailPingMonitor) Name() string {
	return m.name
}

func (m *emailPingMonitor) Type() string {
	return "email_ping"
}

func (m *emailPingMonitor) HostURL() string {
	return m.host
}

func (m *emailPingMonitor) Key() string {
	return m.key
}

func (m *emailPingMonitor) Interval() int {
	return m.interval
}

func (m *emailPingMonitor) Run() (monitorStatus, string, int, error) {
	// Simply let the upstream host know that we are alive
	return StatusUp, "OK", 0, nil
}

// Setup a monitor of type 'email_ping'
func setupEmailPingMonitor(host string, monitor *config.Monitor) *emailPingMonitor {
	//? SMTP
	// region parameter checks
	if monitor.SMTPHost == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "smtp_host",
		)
	}

	if monitor.SMTPPort == 0 {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "smtp_port",
		)
	}

	if monitor.SMTPSenderAddress == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "smtp_sender_address",
		)
	}

	if monitor.SMTPRecipientAddress == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "smtp_recipient_address",
		)
	}

	if monitor.SMTPUsername == "" {
		zap.S().Debugw("Empty paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "smtp_username",
		)
	}

	if monitor.SMTPPassword == "" {
		zap.S().Debugw("Empty paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "smtp_password",
		)
	}

	//? IMAP
	if monitor.IMAPHost == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "imap_host",
		)
	}

	if monitor.IMAPPort == 0 {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "imap_port",
		)
	}

	if monitor.IMAPUsername == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "imap_username",
		)
	}

	if monitor.IMAPPassword == "" {
		zap.S().Debugw("Empty paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "imap_password",
		)
	}

	//? Message
	if monitor.MessageSubject == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "message_subject",
		)
	}

	if monitor.MessageBody == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "message_body",
		)
	}
	if monitor.ResponseSubject == "" {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "response_subject",
		)
	}

	//? Misc
	if monitor.Timeout == 0 {
		zap.S().Panicw("Missing paramter for monitor",
			"name", monitor.Name,
			"type", monitor.Type,
			"paramter", "timeout",
		)
	}
	// endregion

	return &emailPingMonitor{
		name:     monitor.Name,
		host:     host,
		interval: monitor.Interval,
		key:      monitor.Key,

		smtp_host:              monitor.SMTPHost,
		smtp_port:              monitor.SMTPPort,
		smtp_force_tls:         monitor.SMTPForceTLS,
		smtp_sender_address:    monitor.SMTPSenderAddress,
		smtp_recipient_address: monitor.SMTPRecipientAddress,
		smtp_username:          monitor.SMTPUsername,
		smtp_password:          monitor.SMTPPassword,

		imap_host:      monitor.IMAPHost,
		imap_port:      monitor.IMAPPort,
		imap_force_tls: monitor.IMAPForceTLS,
		imap_username:  monitor.IMAPUsername,
		imap_password:  monitor.IMAPPassword,

		message_subject:  monitor.MessageSubject,
		message_body:     monitor.MessageBody,
		response_subject: monitor.ResponseSubject,

		timeout: monitor.Timeout,
	}
}
