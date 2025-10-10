package disks

import (
	"aispace/internal/consts"
	"aispace/internal/storage"
	"context"

	"github.com/google/uuid"
)

type DiskRepository interface {
	CreateDisk(disk Disk) error
	DeleteDisk(id uuid.UUID) error
	GetDisks(ctx context.Context) ([]Disk, error)
	GetProjectsByName(ctx context.Context, name string) ([]DiskProject, error)
	GetProjectNameByID(id uuid.UUID) (string, error)
}

type PostgresDiskRepository struct {
	uow storage.UnitOfWork
}

func NewPostgresDiskRepository(uow storage.UnitOfWork) *PostgresDiskRepository {
	return &PostgresDiskRepository{uow: uow}
}

func (p *PostgresDiskRepository) GetDisks(ctx context.Context) ([]Disk, error) {
	email := ctx.Value(consts.ContextEmail).(string)
	query := `
		SELECT d.id, d.name, u.email, u.name, d.size, d.shared, p.name, d.created_at
		FROM disks d
		JOIN users u
		ON u.id = d.owner_id
		JOIN projects p
		ON p.id = d.project_id
		WHERE u.email = $1
	`

	rows, err := p.uow.DB().Queryx(query, email)

	if err != nil {
		return []Disk{}, err
	}
	defer rows.Close()

	var diskList []Disk

	for rows.Next() {
		var disk Disk
		var ownerName, ownerEmail string

		err = rows.Scan(
			&disk.ID,
			&disk.Name,
			&ownerEmail,
			&ownerName,
			&disk.Size,
			&disk.Shared,
			&disk.Project.Name,
			&disk.CreatedAt,
		)

		if err != nil {
			return []Disk{}, err
		}

		disk.Owner = Owner{
			Username: ownerName,
			Email:    ownerEmail,
		}

		diskList = append(diskList, disk)
	}

	return diskList, nil
}

func (p *PostgresDiskRepository) GetProjectNameByID(id uuid.UUID) (string, error) {
	query := `
		SELECT
		    p.name
		FROM
		    projects p
		WHERE
		    p.id = $1
	`
	rows, err := p.uow.DB().Queryx(query, id)

	if err != nil {
		return "", err
	}
	defer rows.Close()

	var projectName string
	if rows.Next() {
		err = rows.Scan(&projectName)
		if err != nil {
			return "", err
		}
	}

	return projectName, nil
}

func (p *PostgresDiskRepository) GetProjectsByName(ctx context.Context, name string) ([]DiskProject, error) {
	email := ctx.Value(consts.ContextEmail).(string)
	query := `
		SELECT DISTINCT
		    p.id,
		    p.name
		FROM
		    projects p
		JOIN
		    users owner_u ON p.owner_id = owner_u.id
		LEFT JOIN
		    project_user_rel pur ON p.id = pur.project_id
		LEFT JOIN
		    users rel_u ON pur.user_id = rel_u.id
		WHERE
			p.name ILIKE $2
			AND
		    (owner_u.email = $1 OR rel_u.email = $1)
	`
	rows, err := p.uow.DB().Queryx(query, email, "%"+name+"%")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectList []DiskProject

	for rows.Next() {
		var project DiskProject

		err = rows.Scan(
			&project.ID,
			&project.Name,
		)
		if err != nil {
			return nil, err
		}

		projectList = append(projectList, project)
	}

	return projectList, nil
}

func (p *PostgresDiskRepository) CreateDisk(disk Disk) error {
	query := `
		INSERT INTO disks (id, name, owner_id, size, shared, project_id, created_at)
		VALUES ($1, $2, (SELECT id FROM users WHERE email = $3), $4, $5, $6, $7)
	`

	_, err := p.uow.DB().Exec(
		query,
		disk.ID,
		disk.Name,
		disk.Owner.Email,
		disk.Size,
		disk.Shared,
		disk.Project.ID,
		disk.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresDiskRepository) DeleteDisk(id uuid.UUID) error {
	query := `
		DELETE FROM disks WHERE id = $1
	`

	_, err := p.uow.DB().Exec(query, id)

	if err != nil {
		return err
	}

	return nil
}

func ProvidePostgresDiskRepository(uow storage.UnitOfWork) DiskRepository {
	return NewPostgresDiskRepository(uow)
}
