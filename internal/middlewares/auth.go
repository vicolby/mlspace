package middlewares

import (
	"aispace/internal/consts"
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func AuthMiddleware(provider *oidc.Provider, config oauth2.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/auth" {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie("access_token")
			if err != nil {
				http.Redirect(w, r, config.AuthCodeURL("state"), http.StatusTemporaryRedirect)
				return
			}

			verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID, SkipClientIDCheck: true})
			idToken, err := verifier.Verify(r.Context(), cookie.Value)
			if err != nil {
				fmt.Printf("Token verification failed: %v\n", err)
				http.SetCookie(w, &http.Cookie{
					Name:   "access_token",
					Value:  "",
					MaxAge: -1,
				})
				http.Redirect(w, r, config.AuthCodeURL("state"), http.StatusTemporaryRedirect)
				return
			}

			var claims struct {
				Email             string
				EmailVerified     bool
				Name              string
				PreferredUsername string
			}
			idToken.Claims(&claims)

			ctx := r.Context()
			ctx = context.WithValue(ctx, consts.ContextEmail, claims.Email)
			ctx = context.WithValue(ctx, consts.ContextUsername, claims.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
