package users

import "net/http"

type AuthHandler struct {
	authService *AuthService
}

func NewHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.authService.Login(r, w)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.authService.Logout(r, w)
}

func ProvideAuthHandler(authService *AuthService) *AuthHandler {
	return NewHandler(authService)
}
