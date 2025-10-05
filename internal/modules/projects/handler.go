package projects

import (
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

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
	cpu_limit, err := strconv.Atoi(r.FormValue("cpu_limit"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ram_limit, err := strconv.Atoi(r.FormValue("ram_limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	storage_limit, err := strconv.Atoi(r.FormValue("storage_limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	command := CreateProjectCommand{
		Name:         r.FormValue("name"),
		Description:  r.FormValue("description"),
		CPULimit:     cpu_limit,
		RAMLimit:     ram_limit,
		StorageLimit: storage_limit,
	}

	var validate *validator.Validate

	validate = validator.New()
	err = validate.Struct(command)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
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
