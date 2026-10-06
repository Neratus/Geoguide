package requests

type LogoutRequest struct {
	Token string
}

type LogoutResponse struct {
	Success bool
}
