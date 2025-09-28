package main

import (
	"aispace/internal"
	"aispace/internal/config"
	"aispace/internal/modules/projects"
	"aispace/internal/modules/users"
	"aispace/internal/storage"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
)

func ProvideServer(cfg *config.Config, r *chi.Mux) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	fx.New(
		fx.Provide(
			config.ProvideConfig,
			internal.ProvideOIDC,
			internal.ProvideOauth2,
			internal.ProvideRouter,
			storage.NewDB,
			storage.NewUnitOfWork,
			users.ProvidePostgresUserRepository,
			users.ProvideAuthService,
			users.ProvideAuthHandler,
			projects.ProvidePostgresProjectRepository,
			projects.ProvideProjectService,
			projects.ProvideProjectHandler,
			internal.NewHandlers,
			ProvideServer,
		),
		fx.Invoke(func(h *internal.Handlers, r *chi.Mux, provider *oidc.Provider, config oauth2.Config) {
			h.SetupRoutes(r, provider)
		}),
		fx.Invoke(func(srv *http.Server, lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							log.Fatalf("Server failed to start: %v", err)
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return srv.Shutdown(ctx)
				},
			})
		}),
	).Run()
}
