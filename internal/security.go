package internal

import (
	"aispace/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func ProvideOIDC(cfg *config.Config) *oidc.Provider {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, fmt.Sprintf("%s/realms/%s", cfg.Auth.KeycloakURL, cfg.Auth.Realm))
	if err != nil {
		log.Fatal(err)
	}

	return provider
}

func ProvideOauth2(cfg *config.Config, provider *oidc.Provider) oauth2.Config {
	oauth2Config := oauth2.Config{
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
		RedirectURL:  cfg.Auth.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return oauth2Config
}
