package internal

import (
	"aispace/internal/config"
	"aispace/internal/middlewares"
	"aispace/internal/modules/projects"
	"aispace/internal/modules/users"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

type Handlers struct {
	cfg            *config.Config
	oauth2Config   oauth2.Config
	authHandler    *users.AuthHandler
	projectHandler *projects.ProjectHandler
}

func NewHandlers(cfg *config.Config, oauth2Config oauth2.Config, authHandler *users.AuthHandler, projectHandler *projects.ProjectHandler) *Handlers {
	return &Handlers{cfg: cfg, oauth2Config: oauth2Config, authHandler: authHandler, projectHandler: projectHandler}
}

func (h *Handlers) SetupRoutes(r *chi.Mux, provider *oidc.Provider) {
	r.Get("/auth", h.authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(provider, h.oauth2Config))
		r.Get("/", h.projectHandler.GetProjects)
		r.Get("/auth/logout", h.authHandler.Logout)
		r.Get("/projects", h.projectHandler.GetProjects)
		r.Post("/projects/create", h.projectHandler.CreateProject)
		r.Get("/projects/{project_id}", h.projectHandler.GetProject)
		r.Get("/projects/{project_id}/add-users", h.projectHandler.GetAvailableUsers)
		r.Post("/projects/{project_id}/add-users", h.projectHandler.AddParticipants)
		r.Delete("/projects/{project_id}/participants/{participant_id}", h.projectHandler.DeleteParticipant)
	})
}

func ProvideRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.CORSMiddleware(&cfg.CORS))

	return r
}
