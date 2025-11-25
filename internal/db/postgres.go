package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/TOPfiIT/auth-service/internal/config"
	"github.com/TOPfiIT/auth-service/internal/models"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db *sql.DB
}

func InitPostgres(cfg *config.Config) *PostgresDB {
	db_opts := cfg.Postgres.GetDatabaseURL()
	db, err := sql.Open("postgres", db_opts)
	if err != nil {
		log.Fatal("Failed to connect to postgres: ", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping postgres: ", err)
	}

	return &PostgresDB{
		db: db,
	}
}

func (p *PostgresDB) CreateCompany(ctx context.Context, name, password string) (*models.Company, error) {
	company := &models.Company{}
	query := `
		INSERT INTO companies (name, password_hash)
		VALUES ($1, $2)
		RETURNING id, name, password, created_at
	`

	if err := p.db.QueryRowContext(ctx, query, name).Scan(
		&company.ID,
		&company.Name,
		&company.Password,
		&company.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("[Postgres] failed to create company: %w", err)
	}

	return company, nil
}

func (p *PostgresDB) GetCompanyByName(ctx context.Context, name string) (*models.Company, error) {
	company := &models.Company{}
	query := `
		SELECT id, name, password_hash, created_at
		FROM companies
		WHERE name = $1
	`

	if err := p.db.QueryRowContext(ctx, query, name).Scan(
		&company.ID,
		&company.Name,
		&company.Password,
		&company.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("[Postgres] get company by name: %w", err)
	}

	return company, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}
