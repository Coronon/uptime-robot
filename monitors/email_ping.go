package monitors

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/coronon/uptime-robot/config"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Main datapoints associated with a single email
type emailData struct {
	from *mail.Address
	to   *mail.Address

	subject string
	body    string
}

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
	// Compose the email
	from := mail.Address{Name: "", Address: m.smtp_sender_address}
	to := mail.Address{Name: "", Address: m.smtp_recipient_address}
	subject := strings.ReplaceAll(m.message_subject, "{UUID}", uuid.New().String())
	body := fmt.Sprintf("This is a test email sent at %v", time.Now().UTC().Format(time.RFC3339))

	data := &emailData{from: &from, to: &to, subject: subject, body: body}
	zap.S().Debugw("Composed email",
		"from", from.Address,
		"to", to.Address,
		"subject", subject,
		"body", body,
	)

	// Clear old, residual responses (useful when not using a UUID in subject)
	zap.S().Debugln("Cleaning old responses...")
	if message, ping, err := m.receive_email(data, false); err != nil {
		message = fmt.Sprintf("error cleaning old responses: %v", message)

		return StatusDown, message, ping, errors.New(message)
	}
	zap.S().Debugln("Cleaned old responses")

	start := time.Now()
	// Send email to PingPong service
	zap.S().Debugln("Sending email...")
	if message, ping, err := m.send_email(data); err != nil {
		return StatusDown, message, ping, err
	}

	// Receive response from PingPong service
	zap.S().Debugln("Waiting for response...")
	if message, ping, err := m.receive_email(data, true); err != nil {
		return StatusDown, message, ping, err
	}
	end := time.Now()

	return StatusUp, "OK", int(end.Sub(start).Seconds()), nil
}

func (m *emailPingMonitor) send_email(data *emailData) (string, int, error) {
	// Connect to the SMTP server
	smtpAddress := fmt.Sprintf("%s:%d", m.smtp_host, m.smtp_port)
	auth := smtp.PlainAuth("", m.smtp_username, m.smtp_password, m.smtp_host)
	conn, err := smtp.Dial(smtpAddress)
	if err != nil {
		message := fmt.Sprintf("failed to connect to SMTP server: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugw("SMTP dialed", "address", smtpAddress)

	// STARTTLS
	if ok, _ := conn.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName: m.smtp_host,
		}
		if err = conn.StartTLS(config); err != nil {
			message := fmt.Sprintf("failed to starttls: %v", err)

			return message, 0, errors.New(message)
		}
		zap.S().Debugln("SMTP STARTTLS completed")
	} else if m.smtp_force_tls {
		zap.S().Debugln("SMTP STARTTLS extension forced but no support")
		return "SMTP STARTTLS extension forced but no support", 0, errors.New("STARTTLS extension forced but no support")
	} else {
		zap.S().Debugln("SMTP continuing with unencrypted connection!")
	}

	// Authenticate
	if err := conn.Auth(auth); err != nil {
		message := fmt.Sprintf("SMTP authentication failed: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("SMTP authenticated")

	// Set the sender and recipient
	if err := conn.Mail(data.from.Address); err != nil {
		message := fmt.Sprintf("failed to set the sender: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("SMTP set MAIL FROM", "address", data.from.Address)
	if err := conn.Rcpt(data.to.Address); err != nil {
		message := fmt.Sprintf("failed to set the recipient: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("SMTP set RCPT TO", "address", data.to.Address)

	// Send the email body
	wc, err := conn.Data()
	if err != nil {
		message := fmt.Sprintf("failed to open data writer: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("SMTP DATA command started")

	msg := []byte(fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		data.to.String(),
		data.from.String(),
		data.subject,
		data.body,
	))
	if _, err := wc.Write(msg); err != nil {
		message := fmt.Sprintf("failed to write email body: %v", err)

		return message, 0, errors.New(message)
	}
	if err := wc.Close(); err != nil {
		message := fmt.Sprintf("failed to finish writing email body: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("SMTP wrote DATA")

	// Close the connection
	if err := conn.Close(); err != nil {
		message := fmt.Sprintf("failed to send email: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("SMTP email accepted")

	return "", 0, nil
}

// Check IMAP server for email that matches `data` and delete it
//
// If `shouldWaitForResponse` is true, this method will wait for a response to
// match `data` or timeout after the configured amount of seconds.
//
// Make sure to call this method before sending the email to remove all residual
// responses.
func (m *emailPingMonitor) receive_email(data *emailData, shouldWaitForResponse bool) (string, int, error) {
	// Connect to the IMAP server
	imapAddress := fmt.Sprintf("%s:%d", m.imap_host, m.imap_port)
	conn, err := client.Dial(imapAddress)
	if err != nil {
		message := fmt.Sprintf("failed to connect to IMAP server: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugw("IMAP dialed", "address", imapAddress)

	// STARTTLS
	if ok, _ := conn.SupportStartTLS(); ok {
		config := &tls.Config{
			ServerName: m.smtp_host,
		}
		if err = conn.StartTLS(config); err != nil {
			message := fmt.Sprintf("failed to starttls: %v", err)

			return message, 0, errors.New(message)
		}
		zap.S().Debugln("IMAP STARTTLS completed")
	} else if m.smtp_force_tls {
		zap.S().Debugln("IMAP STARTTLS capability forced but no support")
		return "IMAP STARTTLS capability forced but no support", 0, errors.New("STARTTLS capability forced but no support")
	} else {
		zap.S().Debugln("IMAP continuing with unencrypted connection!")
	}

	// Login to the IMAP server
	if err := conn.Login(m.imap_username, m.imap_password); err != nil {
		message := fmt.Sprintf("IMAP login failed: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("IMAP authenticated")

	// Select the INBOX mailbox
	_, err = conn.Select("INBOX", false)
	if err != nil {
		message := fmt.Sprintf("failed to select INBOX: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("IMAP selected INBOX")

	// Search for emails from the recipient
	header_criteria := textproto.MIMEHeader{}
	header_criteria.Add("From", m.smtp_recipient_address)
	header_criteria.Add("Subject", strings.ReplaceAll(m.response_subject, "{ORIG_SUBJ}", data.subject))

	criteria := imap.NewSearchCriteria()
	criteria.Header = header_criteria
	uids, err := conn.UidSearch(criteria)
	if err != nil {
		message := fmt.Sprintf("failed to search for emails: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("IMAP built search critera")

	// Wait for the reply email
	waitStartTime := time.Now()
	for len(uids) == 0 && shouldWaitForResponse {
		zap.S().Debugln("IMAP no response found yet, sleeping...")
		// Sleep for 1 second and check again
		time.Sleep(time.Second)

		uids, err = conn.UidSearch(criteria)
		if err != nil {
			message := fmt.Sprintf("failed to search for emails: %v", err)

			return message, 0, errors.New(message)
		}

		if waitTime := time.Now().Sub(waitStartTime).Seconds(); waitTime > float64(m.timeout) {
			message := fmt.Sprintf("timed out waiting for response after %v seconds", waitTime)

			return message, 0, errors.New(message)
		}
	}
	if len(uids) == 0 {
		zap.S().Debugln("IMAP no response found - not waiting")
		
		return "", 0, nil
	}
	zap.S().Debugln("IMAP response found")

	// Fetch the reply email
	if zap.S().Level() == zap.DebugLevel {
		seqSet := new(imap.SeqSet)
		seqSet.AddNum(uids[0])
		items := []imap.FetchItem{imap.FetchEnvelope}
		messages := make(chan *imap.Message, 1)
		done := make(chan error, 1)
		go func() {
			done <- conn.UidFetch(seqSet, items, messages)
		}()
		zap.S().Debugln("IMAP fetched response")

		var replyEmail *imap.Message
		select {
		case msg := <-messages:
			replyEmail = msg
		case err := <-done:
			if err != nil {
				message := fmt.Sprintf("failed to fetch the reply email: %v", err)

				return message, 0, errors.New(message)
			}
		}

		// Print the reply email details
		zap.S().Debugw("IMAP reply email",
			"from", replyEmail.Envelope.From[0].Address(),
			"subject", replyEmail.Envelope.Subject,
			"body", replyEmail.Body,
		)
	}

	// Delete received email
	set := new(imap.SeqSet)
	set.AddNum(uids...)
	if err := conn.UidStore(set, "+FLAGS.SILENT", []interface{}{`\Deleted`}, nil); err != nil {
		message := fmt.Sprintf("failed to set deleted flag: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("IMAP deleted all response emails in INBOX")

	if err := conn.Expunge(nil); err != nil {
		message := fmt.Sprintf("failed to expunge messages: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debugln("IMAP expunged successfully")

	// Close the connection
	if err := conn.Logout(); err != nil {
		message := fmt.Sprintf("failed to receive email: %v", err)

		return message, 0, errors.New(message)
	}
	zap.S().Debug("IMAP response received")

	return "", 0, nil
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
