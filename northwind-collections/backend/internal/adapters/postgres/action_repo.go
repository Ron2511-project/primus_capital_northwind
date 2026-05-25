package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ron/northwind-collections/internal/domain"
)

type ActionRepo struct {
	pool *pgxpool.Pool
}

func NewActionRepo(pool *pgxpool.Pool) *ActionRepo {
	return &ActionRepo{pool: pool}
}

func (r *ActionRepo) Create(ctx context.Context, action domain.CollectionAction) (*domain.CollectionAction, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO collection_actions (id, customer_id, action_type, notes, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		action.ID, action.CustomerID, action.ActionType, action.Notes, action.CreatedBy, action.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &action, nil
}

func (r *ActionRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.CollectionAction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, customer_id, action_type, notes, created_by, created_at
		FROM collection_actions WHERE customer_id = $1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.CollectionAction
	for rows.Next() {
		a, err := scanActionRows(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *ActionRepo) LastByCustomer(ctx context.Context, customerID uuid.UUID) (*domain.CollectionAction, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, action_type, notes, created_by, created_at
		FROM collection_actions WHERE customer_id = $1 ORDER BY created_at DESC LIMIT 1`, customerID)
	a, err := scanActionRow(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return a, err
}

func scanActionRow(row pgx.Row) (*domain.CollectionAction, error) {
	var a domain.CollectionAction
	var t string
	err := row.Scan(&a.ID, &a.CustomerID, &t, &a.Notes, &a.CreatedBy, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	a.ActionType = domain.ActionType(t)
	return &a, nil
}

func scanActionRows(rows pgx.Rows) (domain.CollectionAction, error) {
	var a domain.CollectionAction
	var t string
	err := rows.Scan(&a.ID, &a.CustomerID, &t, &a.Notes, &a.CreatedBy, &a.CreatedAt)
	a.ActionType = domain.ActionType(t)
	return a, err
}
