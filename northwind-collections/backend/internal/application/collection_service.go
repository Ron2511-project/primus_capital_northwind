package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ron/northwind-collections/internal/domain"
	"github.com/ron/northwind-collections/internal/ports"
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
	ErrInvalidInput     = errors.New("invalid input")
)

type CollectionService struct {
	customers ports.CustomerRepository
	invoices  ports.InvoiceRepository
	actions   ports.CollectionActionRepository
	dashboard ports.DashboardRepository
}

func NewCollectionService(
	customers ports.CustomerRepository,
	invoices ports.InvoiceRepository,
	actions ports.CollectionActionRepository,
	dashboard ports.DashboardRepository,
) *CollectionService {
	return &CollectionService{
		customers: customers,
		invoices:  invoices,
		actions:   actions,
		dashboard: dashboard,
	}
}

func (s *CollectionService) GetQueue(ctx context.Context, segment *domain.Segment, priority *domain.PriorityLevel) ([]domain.CollectionQueueItem, error) {
	return s.customers.List(ctx, segment, priority)
}

func (s *CollectionService) GetSummary(ctx context.Context) (domain.DashboardSummary, error) {
	return s.dashboard.GetSummary(ctx)
}

type CustomerDetail struct {
	Customer domain.Customer            `json:"customer"`
	Invoices []domain.Invoice           `json:"invoices"`
	Actions  []domain.CollectionAction  `json:"actions"`
	Queue    domain.CollectionQueueItem `json:"queue_context"`
}

func (s *CollectionService) GetCustomerDetail(ctx context.Context, id uuid.UUID) (*CustomerDetail, error) {
	customer, err := s.customers.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}

	invoices, err := s.invoices.ListByCustomer(ctx, id)
	if err != nil {
		return nil, err
	}
	actions, err := s.actions.ListByCustomer(ctx, id)
	if err != nil {
		return nil, err
	}

	queue, _ := s.customers.List(ctx, nil, nil)
	var ctxItem domain.CollectionQueueItem
	for _, item := range queue {
		if item.Customer.ID == id {
			ctxItem = item
			break
		}
	}
	if ctxItem.Customer.ID == uuid.Nil {
		ctxItem.Customer = *customer
	}

	return &CustomerDetail{
		Customer: *customer,
		Invoices: invoices,
		Actions:  actions,
		Queue:    ctxItem,
	}, nil
}

type CreateActionInput struct {
	CustomerID uuid.UUID
	ActionType domain.ActionType
	Notes      string
	CreatedBy  string
}

func (s *CollectionService) CreateAction(ctx context.Context, input CreateActionInput) (*domain.CollectionAction, error) {
	if input.CustomerID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Notes == "" && input.ActionType != domain.ActionSnooze {
		return nil, ErrInvalidInput
	}
	if input.CreatedBy == "" {
		input.CreatedBy = "finanzas"
	}

	validTypes := map[domain.ActionType]bool{
		domain.ActionCall: true, domain.ActionEmailReminder: true,
		domain.ActionNote: true, domain.ActionPromiseToPay: true,
		domain.ActionEscalation: true, domain.ActionSnooze: true,
	}
	if !validTypes[input.ActionType] {
		return nil, ErrInvalidInput
	}

	customer, err := s.customers.GetByID(ctx, input.CustomerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, ErrCustomerNotFound
	}

	action := domain.CollectionAction{
		ID:         uuid.New(),
		CustomerID: input.CustomerID,
		ActionType: input.ActionType,
		Notes:      input.Notes,
		CreatedBy:  input.CreatedBy,
		CreatedAt:  time.Now().UTC(),
	}
	return s.actions.Create(ctx, action)
}
