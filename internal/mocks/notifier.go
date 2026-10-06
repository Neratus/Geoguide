package mocks

import (
	"context"
	"sync"

	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

type MockNotifier struct {
	sync.RWMutex
	SentMessages []interfaces.NotifyMessage
}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{SentMessages: []interfaces.NotifyMessage{}}
}

func (m *MockNotifier) Send(ctx context.Context, msg interfaces.NotifyMessage) error {
	m.Lock()
	defer m.Unlock()
	m.SentMessages = append(m.SentMessages, msg)
	return nil
}

func (m *MockNotifier) SendTemplate(ctx context.Context, to, templateName string, data interface{}) error {
	return nil
}
func (m *MockNotifier) SendWelcomeEmail(ctx context.Context, to, username string) error { return nil }
func (m *MockNotifier) SendLoginNotification(ctx context.Context, to, username, ip, userAgent string) error {
	return nil
}
func (m *MockNotifier) SendUserBlockedNotification(ctx context.Context, to, username, reason string) error {
	return nil
}
func (m *MockNotifier) SendUserUnblockedNotification(ctx context.Context, to, username string) error {
	return nil
}
func (m *MockNotifier) SendVerificationEmail(ctx context.Context, to, code string) error { return nil }
