package domain

import "github.com/google/uuid"

type EmailTask struct {
	To       string
	Subject  string
	Template string
	Data     map[string]interface{}
}

type Task interface {
	Name() string
	Payload() interface{}
}

func (t EmailTask) Name() string         { return "send_email" }
func (t EmailTask) Payload() interface{} { return t }

type WelcomeEmailTask struct {
	UserID   uuid.UUID
	Email    string
	Username string
}

func (t WelcomeEmailTask) Name() string         { return "send_welcome_email" }
func (t WelcomeEmailTask) Payload() interface{} { return t }

type LoginNotificationTask struct {
	UserID    uuid.UUID
	Email     string
	Username  string
	IP        string
	UserAgent string
}

func (LoginNotificationTask) Name() string {
	return "send_login_notification"
}

func (t LoginNotificationTask) Payload() interface{} { return t }

type UserBlockedTask struct {
	UserID      uuid.UUID
	Email       string
	Username    string
	Reason      string
	ModeratorID uuid.UUID
}

func (UserBlockedTask) Name() string {
	return "send_user_blocked_notification"
}

func (t UserBlockedTask) Payload() interface{} {
	return t
}

type UserUnblockedTask struct {
	UserID      uuid.UUID
	Email       string
	Username    string
	ModeratorID uuid.UUID
}

func (UserUnblockedTask) Name() string {
	return "send_user_unblocked_notification"
}

func (t UserUnblockedTask) Payload() interface{} {
	return t
}

type NewReviewTask struct {
	ReviewID  uuid.UUID
	PlaceID   uuid.UUID
	PlaceName string
	UserID    uuid.UUID
	Username  string
	Rating    int
	Comment   string
}

func (t NewReviewTask) Name() string {
	return "notify_moderators_new_review"
}

func (t NewReviewTask) Payload() interface{} {
	return t
}

type ReviewModeratedTask struct {
	ReviewID  uuid.UUID
	UserID    uuid.UUID
	UserEmail string
	Username  string
	PlaceName string
	Approved  bool
	Comment   string
}

func (t ReviewModeratedTask) Name() string {
	return "notify_user_review_moderated"
}

func (t ReviewModeratedTask) Payload() interface{} {
	return t
}

type SendVerificationCodeTask struct {
	UserID  uuid.UUID `json:"userId"`
	Contact string    `json:"contact"`
	Purpose string    `json:"purpose"`
	Code    string    `json:"code"`
}

func (t SendVerificationCodeTask) Name() string {
	return "send_verification_code"
}

func (t SendVerificationCodeTask) Payload() interface{} {
	return t
}
