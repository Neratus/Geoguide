package interfaces

type AuthProvider interface {
	GetUserInfo(authCode string) (AuthUser, error)
}
type AuthUser struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}
