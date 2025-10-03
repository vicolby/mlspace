package projects

import (
	"aispace/internal/consts"
	"aispace/web/components"
	"aispace/web/pages/projectsweb"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectService struct {
	repository ProjectRepository
}

func NewProjectService(repository ProjectRepository) *ProjectService {
	return &ProjectService{repository: repository}
}

func (s *ProjectService) GetProjects(w http.ResponseWriter, r *http.Request) http.HandlerFunc {

	projects, err := s.repository.GetProjects(r.Context())

	if err != nil {
		log.Printf("Error while fetching projects: %s", err)
		w.Header().Set("HX-Retarget", "#popup-message")
		w.Header().Set("HX-Reswap", "outerHTML")
		return templ.Handler(components.ErrorPopup("Error fetching projects"), templ.WithStatus(http.StatusInternalServerError)).ServeHTTP
	}

	var webProjectList []projectsweb.WebProject

	for _, project := range projects {
		webProjectList = append(webProjectList, project.ToWebProject(project))
	}

	if r.Header.Get("HX-Request") == "true" {
		return templ.Handler(projectsweb.ProjectsPartial(webProjectList, projectsweb.ModalErrors{})).ServeHTTP
	}
	return templ.Handler(projectsweb.ProjectsFull(webProjectList, projectsweb.ModalErrors{})).ServeHTTP
}

func (s *ProjectService) CreateProject(w http.ResponseWriter, r *http.Request, command CreateProjectCommand) http.HandlerFunc {

	email := r.Context().Value(consts.ContextEmail).(string)
	username := r.Context().Value(consts.ContextUsername).(string)
	projectId := uuid.New()

	project := Project{
		ID:           projectId,
		Name:         command.Name,
		Description:  command.Description,
		Owner:        Owner{Email: email, Username: username},
		CPULimit:     command.CPULimit,
		RAMLimit:     command.RAMLimit,
		StorageLimit: command.StorageLimit,
	}

	err := s.repository.CreateProject(project)

	if err != nil {
		w.Header().Set("HX-Retarget", "#popup-message")
		w.Header().Set("HX-Reswap", "outerHTML")
		return templ.Handler(components.ErrorPopup("Error while creating a project"), templ.WithStatus(http.StatusInternalServerError)).ServeHTTP
	}

	webProject := project.ToWebProject(project)

	return templ.Handler(projectsweb.ProjectRow(webProject)).ServeHTTP
}

func (s *ProjectService) GetProject(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	projectId, err := uuid.Parse(chi.URLParam(r, "project_id"))

	if err != nil {
		log.Printf("Could not parse the project id url param: %s", err)
		w.Header().Set("HX-Retarget", "#popup-message")
		w.Header().Set("HX-Reswap", "outerHTML")
		return templ.Handler(components.ErrorPopup("Project id must be a uuid string"), templ.WithStatus(http.StatusBadRequest)).ServeHTTP
	}

	project, err := s.repository.GetProject(projectId)

	if err != nil {
		log.Printf("Could not fetch project data: %s", err)
		w.Header().Set("HX-Retarget", "#popup-message")
		w.Header().Set("HX-Reswap", "outerHTML")
		return templ.Handler(components.ErrorPopup("Error while fetching the project data"), templ.WithStatus(http.StatusBadRequest)).ServeHTTP
	}
	webProject := project.ToWebProject(*project)

	participants, err := s.repository.GetProjectParticipants(projectId)

	if err != nil {
		log.Printf("Could not fetch project participants data: %s", err)
		w.Header().Set("HX-Retarget", "#popup-message")
		w.Header().Set("HX-Reswap", "outerHTML")
		return templ.Handler(components.ErrorPopup("Error while fetching project participants"), templ.WithStatus(http.StatusBadRequest)).ServeHTTP
	}

	var webParticipants []projectsweb.WebProjectParticipant
	for _, participant := range participants {
		webParticipants = append(webParticipants, participant.ToWebParticipant(participant))
	}

	if r.Header.Get("HX-Request") == "true" {
		return templ.Handler(projectsweb.ProjectPagePartial(webProject, webParticipants)).ServeHTTP
	}

	return templ.Handler(projectsweb.ProjectPageFull(webProject, webParticipants)).ServeHTTP
}

func (s *ProjectService) GetAvailableUsers(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	projectId, err := uuid.Parse(chi.URLParam(r, "project_id"))
	var webParticipants []projectsweb.WebProjectParticipant

	if err != nil {
		log.Printf("Could not parse the project id url param: %s", err)
		w.Header().Set("HX-Retarget", "#popup-message")
		w.Header().Set("HX-Reswap", "outerHTML")
		return templ.Handler(components.ErrorPopup("Project id must be a uuid string"), templ.WithStatus(http.StatusBadRequest)).ServeHTTP
	}

	availableUsers := s.repository.GetAvailableUsers(projectId)

	if availableUsers == nil {
		availableUsers = []Participant{}
	}

	for _, participant := range availableUsers {
		webParticipants = append(webParticipants, participant.ToWebParticipant(participant))
	}

	return templ.Handler(projectsweb.ParticipantModal(webParticipants)).ServeHTTP
}

func (s *ProjectService) AddParticipants(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	var participantUUIDs []uuid.UUID
	r.ParseForm()
	for key := range r.Form {
		if id, err := uuid.Parse(key); err == nil {
			participantUUIDs = append(participantUUIDs, id)
		}
	}

	projectId, err := uuid.Parse(chi.URLParam(r, "project_id"))
	var webParticipants []projectsweb.WebProjectParticipant

	if err != nil {
		log.Printf("Could not parse the project id url param: %s", err)
		w.Header().Set("HX-Retarget", "#popup-message")
		w.Header().Set("HX-Reswap", "outerHTML")
		return templ.Handler(components.ErrorPopup("Project id must be a uuid string"), templ.WithStatus(http.StatusBadRequest)).ServeHTTP
	}

	s.repository.AddParticipants(participantUUIDs, projectId)

	participants, err := s.repository.GetProjectParticipants(projectId)

	for _, participant := range participants {
		webParticipants = append(webParticipants, participant.ToWebParticipant(participant))
	}

	return templ.Handler(projectsweb.ParticipantRows(webParticipants)).ServeHTTP
}

func ProvideProjectService(repository ProjectRepository) *ProjectService {
	return NewProjectService(repository)
}
