package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/gomail.v2"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type SMTPSender struct {
	from        string
	password    string
	host        string
	port        int
	templateDir string
}

func NewSMTPSender(from, password, host string, port int, templateDir string) (*SMTPSender, error) {
	fmt.Println("Initializing SMTP sender with config:", from, host, port, templateDir)
	return &SMTPSender{
		from:        from,
		password:    password,
		host:        host,
		port:        port,
		templateDir: templateDir,
	}, nil
}

func (s *SMTPSender) Send(ctx context.Context, msg interfaces.NotifyMessage) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", msg.To)
	m.SetHeader("Subject", msg.Subject)
	if msg.IsHTML {
		m.SetBody("text/html", msg.Body)
	} else {
		m.SetBody("text/plain", msg.Body)
	}

	dialer := gomail.NewDialer(s.host, s.port, s.from, s.password)
	if s.port == 465 {
		dialer.SSL = true
	}
	if err := dialer.DialAndSend(m); err != nil {
		slog.Error("failed to send email", "error", err)
		return err
	}
	slog.Info("email sent successfully", "to", msg.To)
	return nil
}

func (s *SMTPSender) SendTemplate(ctx context.Context, to, templateName string, data interface{}) error {
	tmplPath := filepath.Join(s.templateDir, templateName+".html")
	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		return err
	}
	tmpl, err := template.New(templateName).Parse(string(tmplBytes))
	if err != nil {
		return err
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}
	return s.Send(ctx, interfaces.NotifyMessage{
		To:      to,
		Subject: s.subjectFromTemplate(templateName),
		Body:    body.String(),
		IsHTML:  true,
	})
}

func (e *SMTPSender) SendWelcomeEmail(ctx context.Context, to, username string) error {
	return e.SendTemplate(ctx, to, "welcome", map[string]interface{}{
		"Username": username,
	})
}

func (s *SMTPSender) subjectFromTemplate(templateName string) string {
	switch templateName {
	case "welcome":
		return "Добро пожаловать в GeoGuide!"
	case "review_moderation":
		return "Ваш отзыв прошел модерацию"
	case "password_reset":
		return "Восстановление пароля"
	case "login_notification":
		return "Уведомление о входе в аккаунт"
	default:
		return "Уведомление от GeoGuide"
	}
}

func (e *SMTPSender) SendUserBlockedNotification(ctx context.Context, to, username, reason string) error {
	return e.SendTemplate(ctx, to, "user_blocked", map[string]interface{}{
		"Username": username,
		"Reason":   reason,
		"Time":     time.Now().Format("02.01.2006 15:04:05"),
	})
}

func (e *SMTPSender) SendUserUnblockedNotification(ctx context.Context, to, username string) error {
	return e.SendTemplate(ctx, to, "user_unblocked", map[string]interface{}{
		"Username": username,
		"Time":     time.Now().Format("02.01.2006 15:04:05"),
	})
}

func (s *SMTPSender) SendLoginNotification(ctx context.Context, to, username, ip, userAgent string) error {
	return s.SendTemplate(ctx, to, "login_notification", map[string]interface{}{
		"Username":  username,
		"IP":        ip,
		"UserAgent": userAgent,
		"Time":      time.Now().Format("02.01.2006 15:04:05"),
	})
}

func (s *SMTPSender) SendVerificationEmail(ctx context.Context, to, code string) error {
	slog.Info("sending verification email", "to", to, "code", code)
	err := s.SendTemplate(ctx, to, "verify_email", map[string]interface{}{
		"Code": code,
		"TTL":  "15 minutes",
	})
	if err != nil {
		slog.Error("failed to send verification email", "error", err)
	}
	return err
}
