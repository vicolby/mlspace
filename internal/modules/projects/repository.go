package projects

import (
	"aispace/internal/consts"
	"aispace/internal/storage"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	GetProjects(ctx context.Context) ([]Project, error)
	GetProject(projectId uuid.UUID) (*Project, error)
	GetProjectParticipants(projectId uuid.UUID) ([]Participant, error)
	GetAvailableUsers(projectId uuid.UUID) []Participant
	CreateProject(project Project) error
	AddParticipants(participants []uuid.UUID, projectId uuid.UUID) error
	DeleteParticipant(participant uuid.UUID, projectId uuid.UUID) error
	CanGetProject(projectId uuid.UUID, ctx context.Context) bool
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
		SELECT DISTINCT
		    p.id,
		    p.name,
		    p.description,
		    owner_u.name AS owner_name,
		    owner_u.email AS owner_email,
		    p.cpu_limit,
		    p.ram_limit,
		    p.storage_limit,
			p.created_at
		FROM
		    projects p
		JOIN
		    users owner_u ON p.owner_id = owner_u.id
		LEFT JOIN
		    project_user_rel pur ON p.id = pur.project_id
		LEFT JOIN
		    users rel_u ON pur.user_id = rel_u.id
		WHERE
		    owner_u.email = $1 OR rel_u.email = $1
		ORDER BY p.created_at DESC
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
			&project.CreatedAt,
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
		SELECT
			projects.id,
			projects.name,
			projects.description,
			users.name,
			users.email,
			projects.cpu_limit,
			projects.ram_limit,
			projects.storage_limit,
			projects.created_at
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
			&project.CreatedAt,
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

func (p *PostgresProjectRepository) GetAvailableUsers(projectId uuid.UUID) []Participant {
	query := `
		SELECT u.id, u.name, u.email
		FROM users u
		LEFT JOIN project_user_rel pur
		ON pur.user_id = u.id
		AND pur.project_id = $1
		WHERE pur.user_id is NULL
		AND u.id != (select p.owner_id from projects p where p.id = $1)
	`
	rows, err := p.uow.DB().Queryx(query, projectId)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var participants []Participant
	for rows.Next() {
		var participant Participant
		err = rows.StructScan(&participant)
		if err != nil {
			return nil
		}
		participants = append(participants, participant)
	}

	return participants
}

func (p *PostgresProjectRepository) AddParticipants(participants []uuid.UUID, projectId uuid.UUID) error {
	query := "INSERT INTO project_user_rel (user_id, project_id) VALUES"
	var args []interface{}
	for i, participant := range participants {
		query += fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args = append(args, participant, projectId)
		if i < len(participants)-1 {
			query += ", "
		}
	}
	query += " ON CONFLICT DO NOTHING"

	_, err := p.uow.DB().Exec(query, args...)

	if err != nil {
		return err
	}

	return nil

}

func (p *PostgresProjectRepository) DeleteParticipant(participant uuid.UUID, projectId uuid.UUID) error {
	query := `
		DELETE FROM project_user_rel pur
		WHERE pur.user_id = $1
		AND pur.project_id = $2
	`
	_, err := p.uow.DB().Exec(query, participant, projectId)

	if err != nil {
		return err
	}

	return nil

}

func (p *PostgresProjectRepository) CanGetProject(projectId uuid.UUID, ctx context.Context) bool {
	email := ctx.Value(consts.ContextEmail).(string)
	query := `
		SELECT DISTINCT
		    p.id,
		    p.name,
		    p.description,
		    owner_u.name AS owner_name,
		    owner_u.email AS owner_email,
		    p.cpu_limit,
		    p.ram_limit,
		    p.storage_limit,
			p.created_at
		FROM
		    projects p
		JOIN
		    users owner_u ON p.owner_id = owner_u.id
		LEFT JOIN
		    project_user_rel pur ON p.id = pur.project_id
		LEFT JOIN
		    users rel_u ON pur.user_id = rel_u.id
		WHERE
			p.id = $2
			AND
		    (owner_u.email = $1 OR rel_u.email = $1)
		ORDER BY p.created_at DESC
	`
	rows, err := p.uow.DB().Queryx(query, email, projectId)

	if err != nil {
		return false
	}

	defer rows.Close()

	return rows.Next()
}

func ProvidePostgresProjectRepository(uow storage.UnitOfWork) ProjectRepository {
	return NewPostgresProjectRepository(uow)
}
