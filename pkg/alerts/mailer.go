package alerts

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/jeffspahr/bourbontracker/pkg/tracker"
)

//go:embed templates/*.html templates/*.txt
var templateFS embed.FS

// Mailer handles sending email alerts
type Mailer struct {
	smtpHost   string
	smtpPort   string
	smtpUser   string
	smtpPass   string
	fromEmail  string
	fromName   string
	htmlTmpl   *template.Template
	textTmpl   *template.Template
}

// EmailData represents the data passed to email templates
type EmailData struct {
	SubscriberID string
	TotalChanges int
	Items        []tracker.InventoryItem
	Timestamp    string
}

// NewMailer creates a new Mailer instance
func NewMailer(fromEmail, fromName string) (*Mailer, error) {
	// Load SMTP config from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	if smtpHost == "" || smtpPort == "" {
		return nil, fmt.Errorf("SMTP configuration missing (SMTP_HOST and SMTP_PORT required)")
	}

	// Load email templates
	htmlTmpl, err := template.ParseFS(templateFS, "templates/new_allocations.html")
	if err != nil {
		return nil, fmt.Errorf("failed to load HTML template: %w", err)
	}

	textTmpl, err := template.ParseFS(templateFS, "templates/new_allocations.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to load text template: %w", err)
	}

	// Use authenticated user as sender if fromEmail not specified
	if fromEmail == "" && smtpUser != "" {
		fromEmail = smtpUser
	}
	if fromEmail == "" {
		fromEmail = "alerts@caskwatch.com"
	}

	return &Mailer{
		smtpHost:  smtpHost,
		smtpPort:  smtpPort,
		smtpUser:  smtpUser,
		smtpPass:  smtpPass,
		fromEmail: fromEmail,
		fromName:  fromName,
		htmlTmpl:  htmlTmpl,
		textTmpl:  textTmpl,
	}, nil
}

// SendAlert sends an alert email to a subscriber
func (m *Mailer) SendAlert(subscriber Subscriber, items []tracker.InventoryItem) error {
	if len(items) == 0 {
		return nil // Nothing to send
	}

	// Prepare template data
	data := EmailData{
		SubscriberID: subscriber.ID,
		TotalChanges: len(items),
		Items:        items,
		Timestamp:    time.Now().Format(time.RFC1123),
	}

	// Render HTML body
	var htmlBody bytes.Buffer
	if err := m.htmlTmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	// Render plain text body (fallback)
	var textBody bytes.Buffer
	if err := m.textTmpl.Execute(&textBody, data); err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	// Build email message with multipart MIME (HTML + text fallback)
	subject := fmt.Sprintf("Cask Watch Alert: %d New Allocation Item%s",
		len(items), pluralize(len(items)))

	msg := m.buildMIMEMessage(subscriber.Email, subject, htmlBody.String(), textBody.String())

	// Send via SMTP
	if err := m.send(subscriber.Email, msg); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", subscriber.Email, err)
	}

	log.Printf("Sent alert to %s (%d items)", subscriber.Email, len(items))
	return nil
}

// SendAlertBatch sends alerts to multiple subscribers with rate limiting
func (m *Mailer) SendAlertBatch(subscribers []Subscriber, itemsPerSubscriber map[string][]tracker.InventoryItem) error {
	sentCount := 0
	errorCount := 0

	for _, sub := range subscribers {
		items, ok := itemsPerSubscriber[sub.ID]
		if !ok || len(items) == 0 {
			continue // No items for this subscriber
		}

		if err := m.SendAlert(sub, items); err != nil {
			log.Printf("ERROR: Failed to send alert to %s: %v", sub.Email, err)
			errorCount++
			continue
		}

		sentCount++

		// Rate limiting: 1 second delay between emails to avoid SMTP throttling
		if sentCount < len(subscribers) {
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("Alert batch complete: %d sent, %d errors", sentCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d email(s) failed to send", errorCount)
	}

	return nil
}

// buildMIMEMessage constructs a multipart MIME message with HTML and text alternatives
func (m *Mailer) buildMIMEMessage(to, subject, htmlBody, textBody string) []byte {
	boundary := "boundary-caskwatch-alert"

	from := m.fromEmail
	if m.fromName != "" {
		from = fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail)
	}

	headers := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/alternative; boundary=\"%s\"\r\n"+
			"\r\n",
		from, to, subject, boundary,
	)

	textPart := fmt.Sprintf(
		"--%s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n"+
			"\r\n",
		boundary, textBody,
	)

	htmlPart := fmt.Sprintf(
		"--%s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n"+
			"\r\n",
		boundary, htmlBody,
	)

	footer := fmt.Sprintf("--%s--\r\n", boundary)

	return []byte(headers + textPart + htmlPart + footer)
}

// send sends an email via SMTP
func (m *Mailer) send(to string, msg []byte) error {
	addr := m.smtpHost + ":" + m.smtpPort

	var auth smtp.Auth
	if m.smtpUser != "" && m.smtpPass != "" {
		auth = smtp.PlainAuth("", m.smtpUser, m.smtpPass, m.smtpHost)
	}

	if err := smtp.SendMail(addr, auth, m.fromEmail, []string{to}, msg); err != nil {
		return err
	}

	return nil
}

// pluralize returns "s" if count != 1, otherwise empty string
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
