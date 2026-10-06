package interfaces

import (
	"context"
)

type NotifyMessage struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}

type Notifier interface {
	Send(ctx context.Context, msg NotifyMessage) error
	SendTemplate(ctx context.Context, to, templateName string, data interface{}) error
	SendWelcomeEmail(ctx context.Context, to, username string) error
	SendLoginNotification(ctx context.Context, to, username, ip, userAgent string) error
	SendUserBlockedNotification(ctx context.Context, to, username, reason string) error
	SendUserUnblockedNotification(ctx context.Context, to, username string) error
	SendVerificationEmail(ctx context.Context, to, code string) error
}
