package users

import (
	"aispace/internal/storage"
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUsers(ctx context.Context) []User
	CreateUser(user User) error
}

type PostgresUserRepository struct {
	uow storage.UnitOfWork
}

func NewPostgresUserRepository(uow storage.UnitOfWork) *PostgresUserRepository {
	return &PostgresUserRepository{uow: uow}
}

func (p *PostgresUserRepository) CreateUser(user User) error {
	query := `
		INSERT INTO users (id, name, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING
	`
	_, err := p.uow.DB().Exec(query, uuid.New(), user.Name, user.Email, time.Now(), time.Now())

	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresUserRepository) GetUsers(ctx context.Context) []User {
	query := `
		SELECT * from users
	`
	rows, err := p.uow.DB().Queryx(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var projectUsers []User
	for rows.Next() {
		var user User
		err = rows.StructScan(&user)
		if err != nil {
			return nil
		}
		projectUsers = append(projectUsers, user)
	}

	return projectUsers
}

func ProvidePostgresUserRepository(uow storage.UnitOfWork) UserRepository {
	return NewPostgresUserRepository(uow)
}
