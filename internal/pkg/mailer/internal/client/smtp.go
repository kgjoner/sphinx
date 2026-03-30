package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/kgjoner/cornucopia/v3/prim"
)

type Client struct {
	auth          smtp.Auth
	addr          string
	host          string
	from          string
	allowInsecure bool
}

func New(username, password, host, port string, allowInsecure bool) *Client {
	smtpUsername := username
	smtpPassword := password
	smtpHost := host
	smtpAddress := smtpHost + ":" + port
	senderAddress := smtpUsername + "@" + smtpHost

	var auth smtp.Auth
	if smtpUsername != "" && smtpPassword != "" {
		auth = smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	}

	return &Client{
		auth:          auth,
		addr:          smtpAddress,
		host:          smtpHost,
		from:          senderAddress,
		allowInsecure: allowInsecure,
	}
}

type SendInput struct {
	To           prim.Email `validate:"required"`
	Subject      string     `validate:"required"`
	Body         string     `validate:"required"`
	AliasName    string
	AliasAddress prim.Email
}

func (m Client) buildMessage(input SendInput) ([]byte, error) {
	if hasHeaderInjection(input.Subject) {
		return nil, fmt.Errorf("invalid subject")
	}

	if hasHeaderInjection(input.AliasName) {
		return nil, fmt.Errorf("invalid alias name")
	}

	// Send email
	senderInfo := m.from
	if input.AliasAddress != "" {
		senderInfo = input.AliasAddress.String()
	}
	if input.AliasName != "" {
		senderInfo = fmt.Sprintf("%v <%v>", input.AliasName, senderInfo)
	}

	msg := fmt.Sprintf("From: %s\r\n", senderInfo)
	msg += fmt.Sprintf("To: %s\r\n", input.To)
	msg += fmt.Sprintf("Subject: %s\r\n", input.Subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "\r\n" + input.Body

	return []byte(msg), nil
}

func (m Client) SendEmail(input SendInput) error {
	msg, err := m.buildMessage(input)
	if err != nil {
		return err
	}

	client, err := smtp.Dial(m.addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if !m.allowInsecure {
		ok, _ := client.Extension("STARTTLS")
		if !ok {
			return errors.New("smtp server does not support STARTTLS; set SMTP_ALLOW_INSECURE=true only for local development")
		}

		tlsCfg := &tls.Config{
			ServerName: m.host,
			MinVersion: tls.VersionTLS12,
		}

		if err = client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	if m.auth != nil {
		if ok, _ := client.Extension("AUTH"); !ok {
			return errors.New("smtp server does not support AUTH")
		}

		if err = client.Auth(m.auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	if err = client.Mail(m.from); err != nil {
		return err
	}

	to := []string{string(input.To)}
	for _, addr := range to {
		if _, err = mail.ParseAddress(addr); err != nil {
			return fmt.Errorf("invalid recipient address %q: %w", addr, err)
		}

		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(msg); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func hasHeaderInjection(value string) bool {
	return strings.Contains(value, "\r") || strings.Contains(value, "\n")
}
