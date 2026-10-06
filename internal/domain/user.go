package domain

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	UserRoleUser      = "USER"
	UserRoleModerator = "MODERATOR"
	UserRoleAnalyst   = "ANALYST"
)

type User struct {
	id                 UserID
	username           string
	email              string
	passwordHash       string
	phone              string
	registeredAt       time.Time
	birthDate          *time.Time
	countryOfResidence string
	avatarURL          string
	role               string
	favoriteCategories []string
	isBlocked          bool
	blockedAt          *time.Time
	blockReason        string

	emailVerified    bool
	phoneVerified    bool
	twoFactorEnabled bool
	twoFactorSecret  string
	backupCodes      []string
}

func (u *User) GetId() UserID                   { return u.id }
func (u *User) GetUsername() string             { return u.username }
func (u *User) GetEmail() string                { return u.email }
func (u *User) GetPasswordHash() string         { return u.passwordHash }
func (u *User) GetPhone() string                { return u.phone }
func (u *User) GetRegisteredAt() time.Time      { return u.registeredAt }
func (u *User) GetBirthDate() *time.Time        { return u.birthDate }
func (u *User) GetCountryOfResidence() string   { return u.countryOfResidence }
func (u *User) GetAvatarUrl() string            { return u.avatarURL }
func (u *User) GetFavoriteCategories() []string { return u.favoriteCategories }
func (u *User) IsBlocked() bool                 { return u.isBlocked }
func (u *User) GetBlockedAt() *time.Time        { return u.blockedAt }
func (u *User) GetBlockReason() string          { return u.blockReason }
func (u *User) GetRole() string                 { return u.role }
func (u *User) IsEmailVerified() bool           { return u.emailVerified }
func (u *User) IsPhoneVerified() bool           { return u.phoneVerified }
func (u *User) IsTwoFactorEnabled() bool        { return u.twoFactorEnabled }
func (u *User) GetTwoFactorSecret() string      { return u.twoFactorSecret }
func (u *User) GetBackupCodes() []string        { return u.backupCodes }

func (u *User) SetId(id UserID)                      { u.id = id }
func (u *User) SetUsername(name string)              { u.username = name }
func (u *User) SetEmail(email string)                { u.email = email }
func (u *User) SetPasswordHash(hash string)          { u.passwordHash = hash }
func (u *User) SetPhone(phone string)                { u.phone = phone }
func (u *User) SetRegisteredAt(t time.Time)          { u.registeredAt = t }
func (u *User) SetBirthDate(d *time.Time)            { u.birthDate = d }
func (u *User) SetCountryOfResidence(country string) { u.countryOfResidence = country }
func (u *User) SetAvatarUrl(url string)              { u.avatarURL = url }
func (u *User) SetFavoriteCategories(cats []string)  { u.favoriteCategories = cats }
func (u *User) SetEmailVerified(v bool)              { u.emailVerified = v }
func (u *User) SetPhoneVerified(v bool)              { u.phoneVerified = v }

func (u *User) EnableTwoFactor(secret string, hashedBackups []string) {
	u.twoFactorEnabled = true
	u.twoFactorSecret = secret
	u.backupCodes = hashedBackups
}
func (u *User) DisableTwoFactor() {
	u.twoFactorEnabled = false
	u.twoFactorSecret = ""
	u.backupCodes = nil
}

func (u *User) Block(reason string) {
	u.isBlocked = true
	now := time.Now()
	u.blockedAt = &now
	u.blockReason = reason
}
func (u *User) Unblock() {
	u.isBlocked = false
	u.blockedAt = nil
	u.blockReason = ""
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	if len(password) < GetConfig().MinPasswordLength {
		return "", ErrPasswordTooShort
	}
	if len(password) > GetConfig().MaxPasswordLength {
		return "", ErrPasswordTooLong
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ComparePassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func GenerateSixDigitCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (u *User) IsModerator() bool   { return u.role == UserRoleModerator }
func (u *User) IsAnalyst() bool     { return u.role == UserRoleAnalyst }
func (u *User) IsRegularUser() bool { return u.role == UserRoleUser }

func NewUser(
	id UserID,
	username, email, passwordHash, phone string,
	birthDate *time.Time,
	countryOfResidence string,
	role string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, ErrEmptyUsername
	}
	if len(username) > GetConfig().MaxUsernameLength {
		return nil, ErrUsernameTooLong
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, ErrEmptyEmail
	}
	if len(email) > GetConfig().MaxEmailLength {
		return nil, ErrEmailTooLong
	}
	if !strings.Contains(email, "@") {
		return nil, ErrInvalidEmailFormat
	}
	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}
	if phone != "" && len(phone) > GetConfig().MaxPhoneLength {
		return nil, ErrPhoneTooLong
	}
	if countryOfResidence != "" && len(countryOfResidence) > GetConfig().MaxCountryResidence {
		return nil, ErrCountryResidenceTooLong
	}
	validRoles := []string{UserRoleUser, UserRoleModerator, UserRoleAnalyst}
	valid := false
	for _, r := range validRoles {
		if role == r {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidUserRole
	}

	return &User{
		id:                 id,
		username:           username,
		email:              email,
		passwordHash:       passwordHash,
		phone:              phone,
		registeredAt:       time.Now(),
		birthDate:          birthDate,
		countryOfResidence: countryOfResidence,
		avatarURL:          "",
		favoriteCategories: []string{},
		isBlocked:          false,
		blockedAt:          nil,
		blockReason:        "",
		role:               role,
	}, nil
}

func NewUserFromDB(
	id UserID,
	username, email, passwordHash, phone string,
	registeredAt time.Time,
	birthDate *time.Time,
	countryOfResidence, avatarURL string,
	favoriteCategories []string,
	isBlocked bool,
	blockedAt *time.Time,
	blockReason string,
	role string,
	emailVerified, phoneVerified, twoFactorEnabled bool,
	twoFactorSecret string,
	backupCodes []string,
) *User {
	return &User{
		id:                 id,
		username:           username,
		email:              email,
		passwordHash:       passwordHash,
		phone:              phone,
		registeredAt:       registeredAt,
		birthDate:          birthDate,
		countryOfResidence: countryOfResidence,
		avatarURL:          avatarURL,
		favoriteCategories: favoriteCategories,
		isBlocked:          isBlocked,
		blockedAt:          blockedAt,
		blockReason:        blockReason,
		role:               role,
		emailVerified:      emailVerified,
		phoneVerified:      phoneVerified,
		twoFactorEnabled:   twoFactorEnabled,
		twoFactorSecret:    twoFactorSecret,
		backupCodes:        backupCodes,
	}
}
