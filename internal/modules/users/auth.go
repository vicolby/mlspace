package users

import (
	"aispace/internal/config"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthService struct {
	repository   UserRepository
	oauth2Config oauth2.Config
	config       *config.Config
	provider     *oidc.Provider
}

var claims struct {
	Email             string
	EmailVerified     bool
	Name              string
	PreferredUsername string
}

func NewAuthService(repository UserRepository, oauth2Config oauth2.Config, config *config.Config, provider *oidc.Provider) *AuthService {
	return &AuthService{repository: repository, oauth2Config: oauth2Config, config: config, provider: provider}
}

func (s *AuthService) Login(r *http.Request, w http.ResponseWriter) {
	if code := r.URL.Query().Get("code"); code != "" {
		token, err := s.oauth2Config.Exchange(r.Context(), code)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)

		if !ok {
			http.Error(w, "No ID token", http.StatusInternalServerError)
			return
		}

		verifier := s.provider.Verifier(&oidc.Config{ClientID: s.oauth2Config.ClientID, SkipClientIDCheck: true})
		idToken, err := verifier.Verify(r.Context(), rawIDToken)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := User{
			ID:        uuid.New(),
			Name:      claims.Name,
			Email:     claims.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = s.repository.CreateUser(user)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if idTokenRaw, ok := token.Extra("id_token").(string); ok {
			http.SetCookie(w, &http.Cookie{
				Name:     "id_token",
				Value:    idTokenRaw,
				Path:     "/",
				MaxAge:   int(time.Until(token.Expiry).Seconds()),
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteLaxMode,
			})
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    token.AccessToken,
			Path:     "/",
			MaxAge:   int(time.Until(token.Expiry).Seconds()),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
}

func (s *AuthService) Logout(r *http.Request, w http.ResponseWriter) {
	idTokenCookie, err := r.Cookie("id_token")
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name: "access_token", Value: "", MaxAge: -1, Path: "/",
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "access_token",
		Value: "",
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "id_token",
		Value: "",
		Path:  "/",
	})

	logoutURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/logout?post_logout_redirect_uri=%s&id_token_hint=%s",
		s.config.Auth.KeycloakURL,
		s.config.Auth.Realm,
		fmt.Sprintf("http://%s:%s", s.config.Server.Host, s.config.Server.Port),
		idTokenCookie.Value,
	)
	fmt.Println(logoutURL)
	http.Redirect(w, r, logoutURL, http.StatusTemporaryRedirect)
}

func ProvideAuthService(repository UserRepository, oauth2Config oauth2.Config, config *config.Config, provider *oidc.Provider) *AuthService {
	return NewAuthService(repository, oauth2Config, config, provider)
}
