package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RunIfEmpty(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM customers`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return seed(ctx, pool)
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	type customerSeed struct {
		name, segment, email string
		terms                int
		mrr                  float64
	}
	customers := []customerSeed{
		{"Acme Corp Global", "enterprise", "ap@acme.com", 75, 12000},
		{"NovaPay Fintech", "startup", "ceo@novapay.io", 30, 2400},
		{"DataStream LLC", "standard", "billing@datastream.com", 30, 890},
		{"GhostMetrics Inc", "zombie", "noreply@ghost.io", 30, 4500},
		{"Helix Retail Group", "enterprise", "finance@helix.com", 75, 15000},
		{"PixelForge Studio", "startup", "hello@pixelforge.dev", 30, 650},
		{"CloudNine SaaS", "standard", "accounts@cloudnine.io", 30, 3200},
		{"StaleAPI Ltd", "zombie", "admin@staleapi.com", 30, 2100},
		{"Meridian Health Tech", "enterprise", "ap@meridian.health", 75, 8500},
		{"QuickLaunch Apps", "startup", "founder@quicklaunch.app", 30, 1200},
		{"OmniLogistics", "standard", "finance@omnilog.com", 30, 5400},
		{"BrightPath Education", "standard", "billing@brightpath.edu", 30, 1800},
	}

	now := time.Now()
	customerIDs := make(map[string]uuid.UUID)

	for _, c := range customers {
		id := uuid.New()
		customerIDs[c.name] = id
		_, err := pool.Exec(ctx, `
			INSERT INTO customers (id, name, segment, payment_terms_days, monthly_mrr, contact_email)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			id, c.name, c.segment, c.terms, c.mrr, c.email)
		if err != nil {
			return fmt.Errorf("insert customer %s: %w", c.name, err)
		}
	}

	type invoiceSeed struct {
		customerName string
		number       string
		amount       float64
		dueDaysAgo   int
		status       string
	}
	invoices := []invoiceSeed{
		{"Acme Corp Global", "INV-2025-001", 12000, 45, "overdue"},
		{"Acme Corp Global", "INV-2025-002", 12000, 15, "pending"},
		{"NovaPay Fintech", "INV-2025-010", 2400, 22, "overdue"},
		{"DataStream LLC", "INV-2025-020", 890, 8, "overdue"},
		{"GhostMetrics Inc", "INV-2024-099", 4500, 120, "overdue"},
		{"GhostMetrics Inc", "INV-2025-050", 4500, 90, "overdue"},
		{"Helix Retail Group", "INV-2025-003", 15000, 60, "overdue"},
		{"PixelForge Studio", "INV-2025-011", 650, 35, "overdue"},
		{"CloudNine SaaS", "INV-2025-021", 3200, 5, "pending"},
		{"StaleAPI Ltd", "INV-2024-080", 2100, 100, "overdue"},
		{"Meridian Health Tech", "INV-2025-004", 8500, 20, "pending"},
		{"QuickLaunch Apps", "INV-2025-012", 1200, 18, "overdue"},
		{"OmniLogistics", "INV-2025-022", 5400, 12, "overdue"},
		{"BrightPath Education", "INV-2025-023", 1800, -5, "pending"},
	}

	for _, inv := range invoices {
		cid := customerIDs[inv.customerName]
		due := now.AddDate(0, 0, -inv.dueDaysAgo)
		_, err := pool.Exec(ctx, `
			INSERT INTO invoices (customer_id, invoice_number, amount, due_date, status)
			VALUES ($1, $2, $3, $4, $5)`,
			cid, inv.number, inv.amount, due, inv.status)
		if err != nil {
			return fmt.Errorf("insert invoice %s: %w", inv.number, err)
		}
	}

	// Acciones de ejemplo
	ghostID := customerIDs["GhostMetrics Inc"]
	_, err := pool.Exec(ctx, `
		INSERT INTO collection_actions (customer_id, action_type, notes, created_by, created_at)
		VALUES ($1, 'email_reminder', 'Recordatorio automático #3 — sin respuesta', 'sistema', $2)`,
		ghostID, now.AddDate(0, 0, -7))
	if err != nil {
		return err
	}

	return nil
}
