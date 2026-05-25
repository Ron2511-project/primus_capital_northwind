package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ron/northwind-collections/internal/domain"
)

// ── Mocks ────────────────────────────────────────────────────────────────────

type mockCustomerRepo struct {
	customers []domain.CollectionQueueItem
	byID      *domain.Customer
	err       error
}

func (m *mockCustomerRepo) List(_ context.Context, _ *domain.Segment, _ *domain.PriorityLevel) ([]domain.CollectionQueueItem, error) {
	return m.customers, m.err
}
func (m *mockCustomerRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Customer, error) {
	return m.byID, m.err
}

type mockInvoiceRepo struct {
	invoices []domain.Invoice
	err      error
}

func (m *mockInvoiceRepo) ListByCustomer(_ context.Context, _ uuid.UUID) ([]domain.Invoice, error) {
	return m.invoices, m.err
}
func (m *mockInvoiceRepo) ListOverdueByCustomer(_ context.Context, _ uuid.UUID) ([]domain.Invoice, error) {
	return m.invoices, m.err
}
func (m *mockInvoiceRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.InvoiceStatus) error {
	return m.err
}

type mockActionRepo struct {
	actions []domain.CollectionAction
	created *domain.CollectionAction
	last    *domain.CollectionAction
	err     error
}

func (m *mockActionRepo) Create(_ context.Context, action domain.CollectionAction) (*domain.CollectionAction, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.created != nil {
		return m.created, nil
	}
	return &action, nil
}
func (m *mockActionRepo) ListByCustomer(_ context.Context, _ uuid.UUID) ([]domain.CollectionAction, error) {
	return m.actions, m.err
}
func (m *mockActionRepo) LastByCustomer(_ context.Context, _ uuid.UUID) (*domain.CollectionAction, error) {
	return m.last, m.err
}

type mockDashboardRepo struct {
	summary domain.DashboardSummary
	err     error
}

func (m *mockDashboardRepo) GetSummary(_ context.Context) (domain.DashboardSummary, error) {
	return m.summary, m.err
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newService(cr *mockCustomerRepo, ir *mockInvoiceRepo, ar *mockActionRepo, dr *mockDashboardRepo) *CollectionService {
	return NewCollectionService(cr, ir, ar, dr)
}

func sampleCustomer(id uuid.UUID) *domain.Customer {
	return &domain.Customer{
		ID:               id,
		Name:             "Acme Corp",
		Segment:          domain.SegmentStandard,
		PaymentTermsDays: 30,
		MonthlyMRR:       5000,
		ContactEmail:     "billing@acme.com",
		IsActive:         true,
		CreatedAt:        time.Now(),
	}
}

// ── GetQueue ──────────────────────────────────────────────────────────────────

func TestGetQueue_ReturnsList(t *testing.T) {
	id := uuid.New()
	items := []domain.CollectionQueueItem{
		{Customer: *sampleCustomer(id), TotalOverdue: 1200, Priority: domain.PriorityMedium},
	}
	svc := newService(&mockCustomerRepo{customers: items}, &mockInvoiceRepo{}, &mockActionRepo{}, &mockDashboardRepo{})

	result, err := svc.GetQueue(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0].Customer.ID != id {
		t.Errorf("expected customer id %s, got %s", id, result[0].Customer.ID)
	}
}

func TestGetQueue_PropagatesRepoError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	svc := newService(&mockCustomerRepo{err: repoErr}, &mockInvoiceRepo{}, &mockActionRepo{}, &mockDashboardRepo{})

	_, err := svc.GetQueue(context.Background(), nil, nil)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

// ── GetCustomerDetail ─────────────────────────────────────────────────────────

func TestGetCustomerDetail_ReturnsFullDetail(t *testing.T) {
	id := uuid.New()
	customer := sampleCustomer(id)
	invoices := []domain.Invoice{
		{ID: uuid.New(), CustomerID: id, Amount: 1500, Status: domain.InvoiceOverdue},
	}
	actions := []domain.CollectionAction{
		{ID: uuid.New(), CustomerID: id, ActionType: domain.ActionCall, Notes: "left voicemail"},
	}
	queueItems := []domain.CollectionQueueItem{
		{Customer: *customer, TotalOverdue: 1500, Priority: domain.PriorityHigh},
	}

	svc := newService(
		&mockCustomerRepo{byID: customer, customers: queueItems},
		&mockInvoiceRepo{invoices: invoices},
		&mockActionRepo{actions: actions},
		&mockDashboardRepo{},
	)

	detail, err := svc.GetCustomerDetail(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Customer.ID != id {
		t.Errorf("wrong customer id")
	}
	if len(detail.Invoices) != 1 {
		t.Errorf("expected 1 invoice, got %d", len(detail.Invoices))
	}
	if len(detail.Actions) != 1 {
		t.Errorf("expected 1 action, got %d", len(detail.Actions))
	}
	if detail.Queue.Priority != domain.PriorityHigh {
		t.Errorf("expected high priority in queue context, got %s", detail.Queue.Priority)
	}
}

func TestGetCustomerDetail_CustomerNotFound(t *testing.T) {
	svc := newService(
		&mockCustomerRepo{byID: nil},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	_, err := svc.GetCustomerDetail(context.Background(), uuid.New())
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("expected ErrCustomerNotFound, got %v", err)
	}
}

func TestGetCustomerDetail_FallsBackWhenNotInQueue(t *testing.T) {
	id := uuid.New()
	customer := sampleCustomer(id)

	svc := newService(
		&mockCustomerRepo{byID: customer, customers: []domain.CollectionQueueItem{}},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	detail, err := svc.GetCustomerDetail(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Queue.Customer.ID != id {
		t.Errorf("expected fallback customer in queue context")
	}
}

// ── CreateAction ──────────────────────────────────────────────────────────────

func TestCreateAction_Success(t *testing.T) {
	id := uuid.New()
	customer := sampleCustomer(id)
	svc := newService(
		&mockCustomerRepo{byID: customer},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	action, err := svc.CreateAction(context.Background(), CreateActionInput{
		CustomerID: id,
		ActionType: domain.ActionCall,
		Notes:      "spoke with CFO",
		CreatedBy:  "agent1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.CustomerID != id {
		t.Errorf("wrong customer id in created action")
	}
	if action.ActionType != domain.ActionCall {
		t.Errorf("wrong action type")
	}
	if action.CreatedBy != "agent1" {
		t.Errorf("wrong created_by")
	}
}

func TestCreateAction_DefaultsCreatedBy(t *testing.T) {
	id := uuid.New()
	svc := newService(
		&mockCustomerRepo{byID: sampleCustomer(id)},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	action, err := svc.CreateAction(context.Background(), CreateActionInput{
		CustomerID: id,
		ActionType: domain.ActionNote,
		Notes:      "some note",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.CreatedBy != "finanzas" {
		t.Errorf("expected default created_by 'finanzas', got '%s'", action.CreatedBy)
	}
}

func TestCreateAction_InvalidCustomerID(t *testing.T) {
	svc := newService(&mockCustomerRepo{}, &mockInvoiceRepo{}, &mockActionRepo{}, &mockDashboardRepo{})

	_, err := svc.CreateAction(context.Background(), CreateActionInput{
		CustomerID: uuid.Nil,
		ActionType: domain.ActionCall,
		Notes:      "test",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateAction_EmptyNotesForNonSnooze(t *testing.T) {
	id := uuid.New()
	svc := newService(
		&mockCustomerRepo{byID: sampleCustomer(id)},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	_, err := svc.CreateAction(context.Background(), CreateActionInput{
		CustomerID: id,
		ActionType: domain.ActionCall,
		Notes:      "   ", // solo espacios
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty notes, got %v", err)
	}
}

func TestCreateAction_SnoozeAllowsEmptyNotes(t *testing.T) {
	id := uuid.New()
	svc := newService(
		&mockCustomerRepo{byID: sampleCustomer(id)},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	_, err := svc.CreateAction(context.Background(), CreateActionInput{
		CustomerID: id,
		ActionType: domain.ActionSnooze,
		Notes:      "",
	})
	if err != nil {
		t.Fatalf("snooze should allow empty notes, got: %v", err)
	}
}

func TestCreateAction_InvalidActionType(t *testing.T) {
	id := uuid.New()
	svc := newService(
		&mockCustomerRepo{byID: sampleCustomer(id)},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	_, err := svc.CreateAction(context.Background(), CreateActionInput{
		CustomerID: id,
		ActionType: "invalid_type",
		Notes:      "some note",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for invalid action type, got %v", err)
	}
}

func TestCreateAction_CustomerNotFound(t *testing.T) {
	id := uuid.New()
	svc := newService(
		&mockCustomerRepo{byID: nil},
		&mockInvoiceRepo{},
		&mockActionRepo{},
		&mockDashboardRepo{},
	)

	_, err := svc.CreateAction(context.Background(), CreateActionInput{
		CustomerID: id,
		ActionType: domain.ActionCall,
		Notes:      "test",
	})
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("expected ErrCustomerNotFound, got %v", err)
	}
}
