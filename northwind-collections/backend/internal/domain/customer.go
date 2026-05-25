package domain

import (
	"time"

	"github.com/google/uuid"
)

type Segment string

const (
	SegmentEnterprise Segment = "enterprise"
	SegmentStartup    Segment = "startup"
	SegmentStandard   Segment = "standard"
	SegmentZombie     Segment = "zombie"
)

type Customer struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Segment          Segment   `json:"segment"`
	PaymentTermsDays int       `json:"payment_terms_days"`
	MonthlyMRR       float64   `json:"monthly_mrr"`
	ContactEmail     string    `json:"contact_email"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}

type Invoice struct {
	ID            uuid.UUID     `json:"id"`
	CustomerID    uuid.UUID     `json:"customer_id"`
	InvoiceNumber string        `json:"invoice_number"`
	Amount        float64       `json:"amount"`
	DueDate       time.Time     `json:"due_date"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	Status        InvoiceStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
}

type InvoiceStatus string

const (
	InvoicePending InvoiceStatus = "pending"
	InvoicePaid    InvoiceStatus = "paid"
	InvoiceOverdue InvoiceStatus = "overdue"
	InvoicePartial InvoiceStatus = "partial"
)

type ActionType string

const (
	ActionCall           ActionType = "call"
	ActionEmailReminder  ActionType = "email_reminder"
	ActionNote           ActionType = "note"
	ActionPromiseToPay   ActionType = "promise_to_pay"
	ActionEscalation     ActionType = "escalation"
	ActionSnooze         ActionType = "snooze"
)

type CollectionAction struct {
	ID         uuid.UUID  `json:"id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	ActionType ActionType `json:"action_type"`
	Notes      string     `json:"notes"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

type PriorityLevel string

const (
	PriorityCritical PriorityLevel = "critical"
	PriorityHigh     PriorityLevel = "high"
	PriorityMedium   PriorityLevel = "medium"
	PriorityLow      PriorityLevel = "low"
	PriorityMonitor  PriorityLevel = "monitor"
)

// CollectionQueueItem agrupa datos para la cola de cobranza.
type CollectionQueueItem struct {
	Customer          Customer       `json:"customer"`
	TotalOverdue      float64        `json:"total_overdue"`
	MaxDaysOverdue    int            `json:"max_days_overdue"`
	OldestDueDate     *time.Time     `json:"oldest_due_date,omitempty"`
	Priority          PriorityLevel  `json:"priority"`
	PriorityScore     int            `json:"priority_score"`
	RecommendedAction string         `json:"recommended_action"`
	OpenInvoiceCount  int            `json:"open_invoice_count"`
	LastActionAt      *time.Time     `json:"last_action_at,omitempty"`
	LastActionType    *ActionType    `json:"last_action_type,omitempty"`
}
