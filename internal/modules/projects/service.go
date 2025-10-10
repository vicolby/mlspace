package projects

import (
	"aispace/internal/base"
	"aispace/internal/clients"
	"aispace/internal/consts"
	"aispace/web/pages/projectsweb"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectService struct {
	repository   ProjectRepository
	kuberService *clients.KuberService
}

func NewProjectService(repository ProjectRepository, kuberService *clients.KuberService) *ProjectService {
	return &ProjectService{repository: repository, kuberService: kuberService}
}

func (s *ProjectService) GetProjects(w http.ResponseWriter, r *http.Request) http.HandlerFunc {

	projects, err := s.repository.GetProjects(r.Context())

	if err != nil {
		log.Printf("Error while fetching projects: %s", err)
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	var webProjectList []projectsweb.WebProject

	for _, project := range projects {
		webProjectList = append(webProjectList, project.ToWebProject(project))
	}

	if r.Header.Get("HX-Request") == "true" {
		return base.Serve(projectsweb.ProjectsPartial(webProjectList), w)
	}
	return base.Serve(projectsweb.ProjectsFull(webProjectList), w)
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

	err := s.kuberService.CreateNamespace(r.Context(), project.ID.String())
	if err != nil {
		log.Println(err)
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	err = s.repository.CreateProject(project)
	if err != nil {
		log.Println(err)
		s.kuberService.DeleteNamespace(r.Context(), project.ID.String())
		return base.ErrorServe("Something went wrong", http.StatusInternalServerError, w)
	}

	webProject := project.ToWebProject(project)

	return base.Serve(projectsweb.ProjectRow(webProject), w)
}

func (s *ProjectService) GetProject(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	projectId, _ := uuid.Parse(chi.URLParam(r, "project_id"))

	if !s.repository.CanGetProject(projectId, r.Context()) {
		return base.ErrorServeRedirect("You can't, brother", http.StatusBadRequest, w)
	}

	project, _ := s.repository.GetProject(projectId)

	webProject := project.ToWebProject(*project)

	participants, _ := s.repository.GetProjectParticipants(projectId)

	var webParticipants []projectsweb.WebProjectParticipant
	for _, participant := range participants {
		webParticipants = append(webParticipants, participant.ToWebParticipant(participant))
	}

	if r.Header.Get("HX-Request") == "true" {
		return base.Serve(projectsweb.ProjectPagePartial(webProject, webParticipants), w)
	}

	return base.Serve(projectsweb.ProjectPageFull(webProject, webParticipants), w)
}

func (s *ProjectService) GetAvailableUsers(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	projectId, err := uuid.Parse(chi.URLParam(r, "project_id"))
	var webParticipants []projectsweb.WebProjectParticipant

	if err != nil {
		return base.ErrorServe("Bad request brother", http.StatusBadRequest, w)
	}

	availableUsers := s.repository.GetAvailableUsers(projectId)

	if availableUsers == nil {
		availableUsers = []Participant{}
	}

	for _, participant := range availableUsers {
		webParticipants = append(webParticipants, participant.ToWebParticipant(participant))
	}

	return base.Serve(projectsweb.ParticipantModal(webParticipants), w)
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
		return base.ErrorServe("Bad request brother", http.StatusBadRequest, w)
	}

	s.repository.AddParticipants(participantUUIDs, projectId)

	participants, err := s.repository.GetProjectParticipants(projectId)

	for _, participant := range participants {
		webParticipants = append(webParticipants, participant.ToWebParticipant(participant))
	}

	return base.Serve(projectsweb.ParticipantRows(webParticipants, true), w)
}

func (s *ProjectService) DeleteParticipant(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	projectId, err := uuid.Parse(chi.URLParam(r, "project_id"))
	participantId, err := uuid.Parse(chi.URLParam(r, "participant_id"))

	if err != nil {
		return base.ErrorServe("Bad request brother", http.StatusBadRequest, w)
	}

	s.repository.DeleteParticipant(participantId, projectId)

	return base.ServeNoSwap(w)
}

func (s *ProjectService) DeleteProject(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	projectId, err := uuid.Parse(chi.URLParam(r, "project_id"))

	if err != nil {
		return base.ErrorServe("Bad request brother", http.StatusBadRequest, w)
	}

	if !s.repository.CanDeleteProject(projectId, r.Context()) {
		return base.ErrorServe("You can't brother", http.StatusBadRequest, w)
	}

	err = s.repository.DeleteProject(projectId)
	if err != nil {
		log.Println(err)
		return base.ErrorServe("Something went wrong", http.StatusBadRequest, w)
	}

	err = s.kuberService.DeleteNamespace(r.Context(), projectId.String())
	if err != nil {
		log.Println(err)
		return base.ErrorServe("Something went wrong", http.StatusBadRequest, w)
	}

	return base.ServeNoSwap(w)
}

func ProvideProjectService(repository ProjectRepository, kuberService *clients.KuberService) *ProjectService {
	return NewProjectService(repository, kuberService)
}
