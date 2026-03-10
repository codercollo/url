package mailer

import (
	"bytes"
	"html/template"
	"log"
	"path/filepath"
	"runtime"

	gomail "gopkg.in/mail.v2"
)

// Config contains SMTP config for sending emails
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Mailer wraps the SMTP dialer and parsed email templates
type Mailer struct {
	dialer *gomail.Dialer
	from   string
	tmpl   *template.Template
}

// New creates a Mailer instance and loads all email templates at startup
func New(cfg Config) (*Mailer, error) {
	//Find the root directory for email template
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "..", "..", "views", "emails")

	//Parse all HTML templates in the email views directory
	tmpl, err := template.ParseGlob(filepath.Join(root, "*.html"))
	if err != nil {
		return nil, err
	}

	//Create SMTP dialer
	d := gomail.NewDialer(
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password)

	return &Mailer{
		dialer: d,
		from:   cfg.From,
		tmpl:   tmpl,
	}, nil

}

// Send renders an email template with data and sends the message
func (m *Mailer) Send(to, subject, templateName string, data any) error {
	//Render the selected template into a buffer
	var body bytes.Buffer
	if err := m.tmpl.ExecuteTemplate(&body, templateName, data); err != nil {
		return err
	}

	//Build the email message
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body.String())

	//Send the email
	if err := m.dialer.DialAndSend(msg); err != nil {
		log.Printf("mailer: fialed to send %q to %s: %v", templateName, to, err)
		return err
	}

	log.Printf("mailer: sent %q to %s", templateName, to)
	return nil
}
