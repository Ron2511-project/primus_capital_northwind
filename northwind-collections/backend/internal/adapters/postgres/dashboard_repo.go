package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ron/northwind-collections/internal/domain"
)

type DashboardRepo struct {
	pool     *pgxpool.Pool
	customer *CustomerRepo
}

func NewDashboardRepo(pool *pgxpool.Pool) *DashboardRepo {
	return &DashboardRepo{pool: pool, customer: NewCustomerRepo(pool)}
}

func (r *DashboardRepo) GetSummary(ctx context.Context) (domain.DashboardSummary, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM customers WHERE is_active = true`).Scan(&total)
	if err != nil {
		return domain.DashboardSummary{}, err
	}

	queue, err := r.customer.List(ctx, nil, nil)
	if err != nil {
		return domain.DashboardSummary{}, err
	}

	var overdueCount, critical, high, zombie int
	var totalOverdue float64
	for _, item := range queue {
		if item.TotalOverdue > 0 {
			overdueCount++
			totalOverdue += item.TotalOverdue
		}
		switch item.Priority {
		case domain.PriorityCritical:
			critical++
		case domain.PriorityHigh:
			high++
		}
		if item.Customer.Segment == domain.SegmentZombie {
			zombie++
		}
	}

	rate := 0.0
	if total > 0 {
		rate = float64(overdueCount) / float64(total) * 100
	}

	return domain.DashboardSummary{
		TotalCustomers:     total,
		OverdueCustomers:   overdueCount,
		TotalOverdueAmount: totalOverdue,
		CriticalCount:      critical,
		HighPriorityCount:  high,
		OverdueRatePercent: rate,
		ZombieCount:        zombie,
	}, nil
}
