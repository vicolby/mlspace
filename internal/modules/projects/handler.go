package projects

import (
	"aispace/web/pages/projectsweb"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	en_translations "github.com/go-playground/validator/v10/translations/en"

	"github.com/a-h/templ"
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

	var (
		uni      *ut.UniversalTranslator
		validate *validator.Validate
	)

	validate = validator.New()
	enLocale := en.New()
	uni = ut.New(enLocale, enLocale)

	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)

	err = validate.Struct(command)

	if err != nil {
		modalErrors := projectsweb.ModalErrors{Errors: make(map[string]string), Values: make(map[string]string)}
		modalErrors.Values["name"] = r.FormValue("name")
		modalErrors.Values["description"] = r.FormValue("description")
		modalErrors.Values["cpu_limit"] = r.FormValue("cpu_limit")
		modalErrors.Values["ram_limit"] = r.FormValue("ram_limit")
		modalErrors.Values["storage_limit"] = r.FormValue("storage_limit")

		var validateErrs validator.ValidationErrors

		if errors.As(err, &validateErrs) {
			for _, err := range validateErrs {
				modalErrors.Errors[err.Field()] = err.Translate(trans)
			}
		}

		w.Header().Set("HX-Retarget", "#new_project_form")
		w.Header().Set("HX-Reswap", "innerHTML")
		templ.Handler(projectsweb.NewProjectForm(modalErrors)).ServeHTTP(w, r)
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

func ProvideProjectHandler(projectService *ProjectService) *ProjectHandler {
	return NewProjectHandler(projectService)
}
