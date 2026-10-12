package services

import (
	"context"
	"encoding/json"

	"github.com/Neratus/geoguide/internal/domain"
	"github.com/Neratus/geoguide/internal/domain/interfaces"
)

func RegisterAllConsumers(queue interfaces.TaskQueue, notifier interfaces.Notifier) error {
	if err := queue.Subscribe(domain.WelcomeEmailTask{}.Name(), func(ctx context.Context, payload []byte) error {
		var task domain.WelcomeEmailTask
		if err := json.Unmarshal(payload, &task); err != nil {
			return err
		}
		return notifier.SendWelcomeEmail(ctx, task.Email, task.Username)
	}); err != nil {
		return err
	}

	if err := queue.Subscribe(domain.LoginNotificationTask{}.Name(), func(ctx context.Context, payload []byte) error {
		var task domain.LoginNotificationTask
		if err := json.Unmarshal(payload, &task); err != nil {
			return err
		}
		return notifier.SendLoginNotification(ctx, task.Email, task.Username, task.IP, task.UserAgent)
	}); err != nil {
		return err
	}

	if err := queue.Subscribe(domain.UserBlockedTask{}.Name(), func(ctx context.Context, payload []byte) error {
		var task domain.UserBlockedTask
		if err := json.Unmarshal(payload, &task); err != nil {
			return err
		}
		return notifier.SendUserBlockedNotification(ctx, task.Email, task.Username, task.Reason)
	}); err != nil {
		return err
	}

	if err := queue.Subscribe(domain.UserUnblockedTask{}.Name(), func(ctx context.Context, payload []byte) error {
		var task domain.UserUnblockedTask
		if err := json.Unmarshal(payload, &task); err != nil {
			return err
		}
		return notifier.SendUserUnblockedNotification(ctx, task.Email, task.Username)
	}); err != nil {
		return err
	}

	if err := queue.Subscribe(domain.SendVerificationCodeTask{}.Name(), func(ctx context.Context, payload []byte) error {
		var task domain.SendVerificationCodeTask
		if err := json.Unmarshal(payload, &task); err != nil {
			return err
		}
		if task.Purpose == "email_verification" {
			return notifier.SendVerificationEmail(ctx, task.Contact, task.Code)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func StartConsumers(ctx context.Context, queue interfaces.TaskQueue) error {
	return queue.Run(ctx)
}
