package storage

import (
	"aispace/internal/config"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type UnitOfWork interface {
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	DB() *sql.DB
}

type unitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (u *unitOfWork) DB() *sql.DB {
	return u.db
}

func NewDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName, cfg.DB.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
