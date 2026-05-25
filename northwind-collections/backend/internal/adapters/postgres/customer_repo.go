package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ron/northwind-collections/internal/domain"
)

type CustomerRepo struct {
	pool *pgxpool.Pool
}

func NewCustomerRepo(pool *pgxpool.Pool) *CustomerRepo {
	return &CustomerRepo{pool: pool}
}

func (r *CustomerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, segment, payment_terms_days, monthly_mrr, contact_email, is_active, created_at
		FROM customers WHERE id = $1`, id)
	return scanCustomer(row)
}

func (r *CustomerRepo) List(ctx context.Context, segment *domain.Segment, priorityFilter *domain.PriorityLevel) ([]domain.CollectionQueueItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name, c.segment, c.payment_terms_days, c.monthly_mrr, c.contact_email, c.is_active, c.created_at
		FROM customers c
		WHERE c.is_active = true
		ORDER BY c.monthly_mrr DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.CollectionQueueItem
	for rows.Next() {
		c, err := scanCustomerRows(rows)
		if err != nil {
			return nil, err
		}
		item, err := r.buildQueueItem(ctx, c)
		if err != nil {
			return nil, err
		}
		if segment != nil && item.Customer.Segment != *segment {
			continue
		}
		if priorityFilter != nil && item.Priority != *priorityFilter {
			continue
		}
		items = append(items, item)
	}
	sortQueue(items)
	return items, rows.Err()
}

func (r *CustomerRepo) buildQueueItem(ctx context.Context, c domain.Customer) (domain.CollectionQueueItem, error) {
	invRows, err := r.pool.Query(ctx, `
		SELECT id, customer_id, invoice_number, amount, due_date, paid_at, status, created_at
		FROM invoices WHERE customer_id = $1 AND status IN ('overdue', 'pending', 'partial')
		ORDER BY due_date ASC`, c.ID)
	if err != nil {
		return domain.CollectionQueueItem{}, err
	}
	defer invRows.Close()

	var totalOverdue float64
	var maxDays int
	var oldest *time.Time
	var openCount int
	var has90 bool
	now := time.Now()

	for invRows.Next() {
		inv, err := scanInvoiceRows(invRows)
		if err != nil {
			return domain.CollectionQueueItem{}, err
		}
		days := int(now.Sub(inv.DueDate).Hours() / 24)
		if inv.Status == domain.InvoiceOverdue || (inv.Status == domain.InvoicePending && days > 0) {
			if days > 0 {
				totalOverdue += inv.Amount
				openCount++
				if days > maxDays {
					maxDays = days
				}
				if oldest == nil || inv.DueDate.Before(*oldest) {
					t := inv.DueDate
					oldest = &t
				}
				if days >= 90 {
					has90 = true
				}
			}
		}
	}

	actionRepo := NewActionRepo(r.pool)
	last, _ := actionRepo.LastByCustomer(ctx, c.ID)

	priority, score, rec := domain.CalculatePriority(
		c.Segment, c.PaymentTermsDays, maxDays, totalOverdue, c.MonthlyMRR, has90, last,
	)

	var lastType *domain.ActionType
	var lastAt *time.Time
	if last != nil {
		lastType = &last.ActionType
		lastAt = &last.CreatedAt
	}

	return domain.CollectionQueueItem{
		Customer:          c,
		TotalOverdue:      totalOverdue,
		MaxDaysOverdue:    maxDays,
		OldestDueDate:     oldest,
		Priority:          priority,
		PriorityScore:     score,
		RecommendedAction: rec,
		OpenInvoiceCount:  openCount,
		LastActionAt:      lastAt,
		LastActionType:    lastType,
	}, nil
}

func sortQueue(items []domain.CollectionQueueItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].PriorityScore > items[i].PriorityScore {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func scanCustomer(row pgx.Row) (*domain.Customer, error) {
	var c domain.Customer
	var seg string
	err := row.Scan(&c.ID, &c.Name, &seg, &c.PaymentTermsDays, &c.MonthlyMRR, &c.ContactEmail, &c.IsActive, &c.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	c.Segment = domain.Segment(seg)
	return &c, nil
}

func scanCustomerRows(rows pgx.Rows) (domain.Customer, error) {
	var c domain.Customer
	var seg string
	err := rows.Scan(&c.ID, &c.Name, &seg, &c.PaymentTermsDays, &c.MonthlyMRR, &c.ContactEmail, &c.IsActive, &c.CreatedAt)
	c.Segment = domain.Segment(seg)
	return c, err
}
