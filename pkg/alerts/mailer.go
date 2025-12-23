package alerts

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/jeffspahr/bourbontracker/pkg/tracker"
	"github.com/mailgun/mailgun-go/v4"
)

//go:embed templates/*.html templates/*.txt
var templateFS embed.FS

// Mailer handles sending email alerts via Mailgun
type Mailer struct {
	mg        *mailgun.MailgunImpl
	fromEmail string
	fromName  string
	htmlTmpl  *template.Template
	textTmpl  *template.Template
}

// EmailData represents the data passed to email templates
type EmailData struct {
	SubscriberID string
	TotalChanges int
	Items        []tracker.InventoryItem
	Timestamp    string
}

// NewMailer creates a new Mailer instance using Mailgun
func NewMailer(fromEmail, fromName string) (*Mailer, error) {
	// Load Mailgun config from environment
	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")

	if mailgunDomain == "" || mailgunAPIKey == "" {
		return nil, fmt.Errorf("Mailgun configuration missing (MAILGUN_DOMAIN and MAILGUN_API_KEY required)")
	}

	// Initialize Mailgun client
	mg := mailgun.NewMailgun(mailgunDomain, mailgunAPIKey)

	// Load email templates
	htmlTmpl, err := template.ParseFS(templateFS, "templates/new_allocations.html")
	if err != nil {
		return nil, fmt.Errorf("failed to load HTML template: %w", err)
	}

	textTmpl, err := template.ParseFS(templateFS, "templates/new_allocations.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to load text template: %w", err)
	}

	// Use default sender if fromEmail not specified
	if fromEmail == "" {
		fromEmail = fmt.Sprintf("postmaster@%s", mailgunDomain)
	}

	return &Mailer{
		mg:        mg,
		fromEmail: fromEmail,
		fromName:  fromName,
		htmlTmpl:  htmlTmpl,
		textTmpl:  textTmpl,
	}, nil
}

// SendAlert sends an alert email to a subscriber via Mailgun
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

	// Build email subject
	subject := fmt.Sprintf("Cask Watch Alert: %d New Allocation Item%s",
		len(items), pluralize(len(items)))

	// Send via Mailgun
	if err := m.send(subscriber.Email, subject, htmlBody.String(), textBody.String()); err != nil {
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

		// Rate limiting: small delay between emails to respect Mailgun limits
		// Mailgun sandbox: 300 emails/day, production varies by plan
		if sentCount < len(subscribers) {
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Printf("Alert batch complete: %d sent, %d errors", sentCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d email(s) failed to send", errorCount)
	}

	return nil
}

// send sends an email via Mailgun API
func (m *Mailer) send(to, subject, htmlBody, textBody string) error {
	// Build from address with name if provided
	from := m.fromEmail
	if m.fromName != "" {
		from = fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail)
	}

	// Create Mailgun message with both HTML and text parts
	message := m.mg.NewMessage(from, subject, textBody, to)
	message.SetHtml(htmlBody)

	// Send with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := m.mg.Send(ctx, message)
	if err != nil {
		return err
	}

	log.Printf("Mailgun message sent with ID: %s", id)
	return nil
}

// pluralize returns "s" if count != 1, otherwise empty string
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
