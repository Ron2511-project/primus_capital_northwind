package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ron/northwind-collections/internal/domain"
)

type CustomerRepository interface {
	List(ctx context.Context, segment *domain.Segment, priority *domain.PriorityLevel) ([]domain.CollectionQueueItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
}

type InvoiceRepository interface {
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Invoice, error)
	ListOverdueByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Invoice, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus) error
}

type CollectionActionRepository interface {
	Create(ctx context.Context, action domain.CollectionAction) (*domain.CollectionAction, error)
	ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.CollectionAction, error)
	LastByCustomer(ctx context.Context, customerID uuid.UUID) (*domain.CollectionAction, error)
}

type DashboardRepository interface {
	GetSummary(ctx context.Context) (domain.DashboardSummary, error)
}
