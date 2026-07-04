// Package mailer
package mailer

import (
	"context"
	"fmt"
	"notifications/internal/config"
	"time"

	gomail "github.com/wneessen/go-mail"
)

type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

type Client struct {
	*gomail.Client
	from     string
	fromName string
}

func New(ctx context.Context, cfg *config.SMTPConfig) (*Client, error) {
	client, err := gomail.NewClient(
		cfg.Host,
		gomail.WithPort(cfg.Port),
		gomail.WithSSL(),
		gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
		gomail.WithUsername(cfg.Username),
		gomail.WithPassword(cfg.Password),
		gomail.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("create mail client: %w", err)
	}

	return &Client{
		Client:   client,
		from:     cfg.From,
		fromName: cfg.FromName,
	}, nil
}

func (c *Client) Send(ctx context.Context, msg Message) error {
	doc := gomail.NewMsg()

	if err := doc.SetAddrHeader(gomail.HeaderFrom, fmt.Sprintf("%s <%s>", c.fromName, c.from)); err != nil {
		return fmt.Errorf("set from header: %w", err)
	}

	if err := doc.AddTo(msg.To); err != nil {
		return fmt.Errorf("set to header: %w", err)
	}

	doc.SetHeader(gomail.HeaderSubject, msg.Subject)
	doc.SetBodyString(gomail.TypeTextPlain, msg.Text)
	doc.AddAlternativeString(gomail.TypeTextHTML, msg.HTML)

	if err := c.Client.DialWithContext(ctx); err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Client.Close()

	if err := c.Client.Send(doc); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	return nil
}
