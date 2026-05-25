package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ron/northwind-collections/internal/domain"
)

type InvoiceRepo struct {
	pool *pgxpool.Pool
}

func NewInvoiceRepo(pool *pgxpool.Pool) *InvoiceRepo {
	return &InvoiceRepo{pool: pool}
}

func (r *InvoiceRepo) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Invoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, customer_id, invoice_number, amount, due_date, paid_at, status, created_at
		FROM invoices WHERE customer_id = $1 ORDER BY due_date DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Invoice
	for rows.Next() {
		inv, err := scanInvoiceRows(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

func (r *InvoiceRepo) ListOverdueByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Invoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, customer_id, invoice_number, amount, due_date, paid_at, status, created_at
		FROM invoices WHERE customer_id = $1 AND status = 'overdue' ORDER BY due_date ASC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Invoice
	for rows.Next() {
		inv, err := scanInvoiceRows(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, rows.Err()
}

func (r *InvoiceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE invoices SET status = $2 WHERE id = $1`, id, status)
	return err
}

func scanInvoiceRows(rows pgx.Rows) (domain.Invoice, error) {
	var inv domain.Invoice
	var st string
	err := rows.Scan(&inv.ID, &inv.CustomerID, &inv.InvoiceNumber, &inv.Amount, &inv.DueDate, &inv.PaidAt, &st, &inv.CreatedAt)
	inv.Status = domain.InvoiceStatus(st)
	return inv, err
}
