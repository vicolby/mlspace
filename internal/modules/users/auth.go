package users

import (
	"aispace/internal/config"
	"aispace/internal/storage"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthService struct {
	uow          storage.UnitOfWork
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

func NewUserService(uow storage.UnitOfWork, oauth2Config oauth2.Config, config *config.Config, provider *oidc.Provider) *AuthService {
	return &AuthService{uow: uow, oauth2Config: oauth2Config, config: config, provider: provider}
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

		query := `INSERT INTO users (id, name, email, role, is_blocked, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`
		_, err = s.uow.DB().Exec(query, uuid.New().String(), claims.Name, claims.Email, "user", false, time.Now(), time.Now())

		if err != nil {
			log.Println(err)
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

func (s *AuthService) GetCurrentUser(r *http.Request) (struct {
	Email    string
	Username string
}, error) {
	cookie, _ := r.Cookie("id_token")
	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.oauth2Config.ClientID, SkipClientIDCheck: true})
	idToken, _ := verifier.Verify(r.Context(), cookie.Value)
	var claims struct {
		Email    string
		Username string
	}
	idToken.Claims(&claims)
	return claims, nil
}

func ProvideAuthService(uow storage.UnitOfWork, oauth2Config oauth2.Config, config *config.Config, provider *oidc.Provider) *AuthService {
	return NewUserService(uow, oauth2Config, config, provider)
}
