package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"car-bridge/internal/entity"
)

type SaferwebCompanyRepository struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewSaferwebCompanyRepository(
	db *pgxpool.Pool,
	log *logrus.Log,
) *SaferwebCompanyRepository {

	return &SaferwebCompanyRepository{
		DB:  db,
		Log: log,
	}

}

func (r *SaferwebCompanyRepository) FindByDOT(
	ctx context.Context,
	dotNumber string,
) (*entity.SaferwebCompany, error) {

	const q = `
		SELECT id, dot_number, legal_name, dba_name, raw_json, created_at, updated_at
		FROM saferweb_company
		WHERE dot_number = $1`

	var e entity.SaferwebCompany
	err := r.DB.QueryRow(
		ctx, q, dotNumber,
	).Scan(
		&e.ID, &e.DOTNumber, &e.LegalName, &e.DBAName,
		&e.RawJSON, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &e, nil

}

func (r *SaferwebCompanyRepository) Upsert(
	ctx context.Context,
	e *entity.SaferwebCompany,
) error {

	const q = `
		INSERT INTO saferweb_company (dot_number, legal_name, dba_name, raw_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		ON CONFLICT (dot_number) DO UPDATE SET
			legal_name = EXCLUDED.legal_name,
			dba_name   = EXCLUDED.dba_name,
			raw_json   = EXCLUDED.raw_json,
			updated_at = now()
		RETURNING id`

	return r.DB.QueryRow(
		ctx, q,
		e.DOTNumber, e.LegalName, e.DBAName, e.RawJSON,
	).Scan(&e.ID)

}
