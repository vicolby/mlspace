package projects

import (
	"net/http"

	"github.com/go-playground/form/v4"
)

var formDecoder *form.Decoder

func init() {
	formDecoder = form.NewDecoder()
}

type ProjectHandler struct {
	projectService *ProjectService
}

func NewProjectHandler(projectService *ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	handler := h.projectService.GetProjects(w, r)
	if handler != nil {
		handler(w, r)
	}
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	handler := h.projectService.GetProject(w, r)
	if handler != nil {
		handler(w, r)
	}
}

func (h *ProjectHandler) GetAvailableUsers(w http.ResponseWriter, r *http.Request) {
	handler := h.projectService.GetAvailableUsers(w, r)
	if handler != nil {
		handler(w, r)
	}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data: "+err.Error(), http.StatusBadRequest)
	}

	command := CreateProjectCommand{}

	if err := formDecoder.Decode(&command, r.PostForm); err != nil {
		http.Error(w, "Invalid input data: "+err.Error(), http.StatusBadRequest)
	}

	if err := command.Validate(); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
	}

	handler := h.projectService.CreateProject(w, r, command)

	if handler != nil {
		handler(w, r)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) AddParticipants(w http.ResponseWriter, r *http.Request) {
	handler := h.projectService.AddParticipants(w, r)
	if handler != nil {
		handler(w, r)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	handler := h.projectService.DeleteParticipant(w, r)
	if handler != nil {
		handler(w, r)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	handler := h.projectService.DeleteProject(w, r)
	if handler != nil {
		handler(w, r)
	} else {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func ProvideProjectHandler(projectService *ProjectService) *ProjectHandler {
	return NewProjectHandler(projectService)
}
