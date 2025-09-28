package projects

import (
	"aispace/internal/middlewares"
	"aispace/internal/modules/users"
	"aispace/internal/storage"
	"aispace/web/pages/projectsweb"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectService struct {
	uow         storage.UnitOfWork
	authService *users.AuthService
}

func NewProjectService(uow storage.UnitOfWork, authService *users.AuthService) *ProjectService {
	return &ProjectService{uow: uow, authService: authService}
}

func (s *ProjectService) GetProjects(r *http.Request) http.HandlerFunc {
	email := r.Context().Value(middlewares.ContextEmail)

	query := `
		SELECT projects.id, projects.name, projects.description, users.name, users.email, projects.cpu_limit, projects.ram_limit, projects.storage_limit 
		FROM projects 
		JOIN users ON projects.owner_id = users.id 
		WHERE users.email = $1
		ORDER BY projects.created_at DESC
	`
	rows, err := s.uow.DB().Query(query, email)

	if err != nil {
		return nil
	}

	defer rows.Close()

	var project_list []projectsweb.WebProject

	for rows.Next() {
		var project projectsweb.WebProject
		err = rows.Scan(&project.ID, &project.Name, &project.Description, &project.OwnerUsername, &project.OwnerEmail, &project.CPULimit, &project.RAMLimit, &project.StorageLimit)

		if err != nil {
			return nil
		}
		project_list = append(project_list, project)
	}

	if r.Header.Get("HX-Request") == "true" {
		return templ.Handler(projectsweb.ProjectsPartial(project_list, projectsweb.ModalErrors{})).ServeHTTP
	}

	return templ.Handler(projectsweb.ProjectsFull(project_list, projectsweb.ModalErrors{})).ServeHTTP
}

func (s *ProjectService) CreateProject(r *http.Request) http.HandlerFunc {

	email := r.Context().Value("email").(string)

	user_query := `SELECT id, name, email FROM users WHERE email = $1`
	rows, err := s.uow.DB().Query(user_query, email)

	if err != nil {
		return nil
	}

	var projectParticipant projectsweb.WebProjectParticipant

	for rows.Next() {
		err = rows.Scan(&projectParticipant.ID, &projectParticipant.Name, &projectParticipant.Email)
		if err != nil {
			return nil
		}
	}

	project_name := r.FormValue("name")
	project_description := r.FormValue("description")
	project_cpu_limit := r.FormValue("cpu_limit")
	project_ram_limit := r.FormValue("ram_limit")
	project_storage_limit := r.FormValue("storage_limit")
	project_id := uuid.New()
	project_query := `INSERT INTO projects (id, name, description, owner_id, cpu_limit, ram_limit, storage_limit) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	rows, err = s.uow.DB().Query(project_query, project_id, project_name, project_description, projectParticipant.ID, project_cpu_limit, project_ram_limit, project_storage_limit)

	if err != nil {
		return nil
	}

	defer rows.Close()

	var web_project projectsweb.WebProject
	web_project.ID = project_id
	web_project.Name = project_name
	web_project.Description = project_description
	web_project.OwnerUsername = projectParticipant.Name
	web_project.OwnerEmail = projectParticipant.Email
	web_project.CPULimit, err = strconv.Atoi(project_cpu_limit)
	if err != nil {
		return nil
	}
	web_project.RAMLimit, err = strconv.Atoi(project_ram_limit)
	if err != nil {
		return nil
	}
	web_project.StorageLimit, err = strconv.Atoi(project_storage_limit)
	if err != nil {
		return nil
	}

	return templ.Handler(projectsweb.ProjectRow(web_project)).ServeHTTP
}

func (s *ProjectService) GetProject(r *http.Request) http.HandlerFunc {
	project_id := chi.URLParam(r, "project_id")
	project_query := `
		SELECT projects.id, projects.name, projects.description, users.name, users.email, projects.cpu_limit, projects.ram_limit, projects.storage_limit 
		FROM projects 
		JOIN users ON projects.owner_id = users.id 
		WHERE projects.id = $1
	`
	rows, err := s.uow.DB().Query(project_query, project_id)
	if err != nil {
		fmt.Println("Error fetching project from DB")
		return nil
	}

	var project projectsweb.WebProject
	for rows.Next() {
		err = rows.Scan(&project.ID, &project.Name, &project.Description, &project.OwnerUsername, &project.OwnerEmail, &project.CPULimit, &project.RAMLimit, &project.StorageLimit)
		if err != nil {
			return nil
		}
	}

	participants_query := `
		SELECT u.id, u.name, u.email
		FROM project_user_rel pur 
		JOIN users u on pur.user_id = u.id
		WHERE pur.project_id = $1
		ORDER BY u.name
	`
	rows, err = s.uow.DB().Query(participants_query, project_id)
	if err != nil {
		fmt.Printf("Error fetching participants %s", err)
		return nil
	}

	var projectParticipants []projectsweb.WebProjectParticipant
	for rows.Next() {
		var participant projectsweb.WebProjectParticipant
		err = rows.Scan(&participant.ID, &participant.Name, &participant.Email)
		if err != nil {
			return nil
		}
		projectParticipants = append(projectParticipants, participant)
	}

	if r.Header.Get("HX-Request") == "true" {
		return templ.Handler(projectsweb.ProjectPagePartial(project, projectParticipants)).ServeHTTP
	}

	return templ.Handler(projectsweb.ProjectPageFull(project, projectParticipants)).ServeHTTP
}

func ProvideProjectService(uow storage.UnitOfWork, authService *users.AuthService) *ProjectService {
	return NewProjectService(uow, authService)
}
