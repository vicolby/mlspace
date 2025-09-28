package projects

import (
	"aispace/internal/consts"
	"aispace/internal/storage"
	"context"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	GetProjects(ctx context.Context) ([]Project, error)
	GetProject(projectId uuid.UUID) (*Project, error)
	GetProjectParticipants(projectId uuid.UUID) ([]Participant, error)
	CreateProject(project Project) error
}

type PostgresProjectRepository struct {
	uow storage.UnitOfWork
}

func NewPostgresProjectRepository(uow storage.UnitOfWork) *PostgresProjectRepository {
	return &PostgresProjectRepository{uow: uow}
}

func (p *PostgresProjectRepository) GetProjects(ctx context.Context) ([]Project, error) {
	email := ctx.Value(consts.ContextEmail).(string)
	query := `
		SELECT projects.id, projects.name, projects.description, users.name, users.email, projects.cpu_limit, projects.ram_limit, projects.storage_limit 
		FROM projects 
		JOIN users ON projects.owner_id = users.id 
		WHERE users.email = $1
		ORDER BY projects.created_at DESC
	`
	rows, err := p.uow.DB().Queryx(query, email)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectList []Project

	for rows.Next() {
		var project Project
		var ownerName, ownerEmail string

		err = rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&ownerName,
			&ownerEmail,
			&project.CPULimit,
			&project.RAMLimit,
			&project.StorageLimit,
		)
		if err != nil {
			return nil, err
		}

		project.Owner = Owner{
			Username: ownerName,
			Email:    ownerEmail,
		}

		projectList = append(projectList, project)
	}

	return projectList, nil
}

func (p *PostgresProjectRepository) GetProject(projectId uuid.UUID) (*Project, error) {
	project_query := `
		SELECT projects.id, projects.name, projects.description, users.name, users.email, projects.cpu_limit, projects.ram_limit, projects.storage_limit 
		FROM projects 
		JOIN users ON projects.owner_id = users.id 
		WHERE projects.id = $1
	`
	rows, err := p.uow.DB().Queryx(project_query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var project Project
	for rows.Next() {
		var ownerName, ownerEmail string

		err = rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&ownerName,
			&ownerEmail,
			&project.CPULimit,
			&project.RAMLimit,
			&project.StorageLimit,
		)
		if err != nil {
			return nil, err
		}

		project.Owner = Owner{
			Username: ownerName,
			Email:    ownerEmail,
		}
	}

	return &project, nil
}

func (p *PostgresProjectRepository) GetProjectParticipants(projectId uuid.UUID) ([]Participant, error) {
	participants_query := `
		SELECT u.id, u.name, u.email
		FROM project_user_rel pur 
		JOIN users u on pur.user_id = u.id
		WHERE pur.project_id = $1
		ORDER BY u.name
	`
	rows, err := p.uow.DB().Queryx(participants_query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectParticipants []Participant
	for rows.Next() {
		var participant Participant
		err = rows.StructScan(&participant)
		if err != nil {
			return nil, err
		}
		projectParticipants = append(projectParticipants, participant)
	}

	return projectParticipants, nil
}

func (p *PostgresProjectRepository) CreateProject(project Project) error {
	query := `
		INSERT INTO projects (id, name, description, owner_id, cpu_limit, ram_limit, storage_limit) 
		VALUES ($1, $2, $3, (SELECT id FROM users WHERE email = $7), $4, $5, $6)
	`
	_, err := p.uow.DB().Exec(
		query,
		project.ID,
		project.Name,
		project.Description,
		project.CPULimit,
		project.RAMLimit,
		project.StorageLimit,
		project.Owner.Email,
	)

	if err != nil {
		return err
	}

	return nil
}

func ProvidePostgresProjectRepository(uow storage.UnitOfWork) ProjectRepository {
	return NewPostgresProjectRepository(uow)
}
