package disks

import (
	"aispace/internal/consts"
	"aispace/internal/storage"
	"context"
)

type DiskRepository interface {
	GetDisks(ctx context.Context) ([]Disk, error)
	GetProjectsByName(ctx context.Context, name string) ([]DiskProject, error)
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
			&disk.Project,
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

func ProvidePostgresDiskRepository(uow storage.UnitOfWork) DiskRepository {
	return NewPostgresDiskRepository(uow)
}
