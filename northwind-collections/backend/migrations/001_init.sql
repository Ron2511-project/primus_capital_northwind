CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE customer_segment AS ENUM ('enterprise', 'startup', 'standard', 'zombie');
CREATE TYPE invoice_status AS ENUM ('pending', 'paid', 'overdue', 'partial');
CREATE TYPE action_type AS ENUM ('call', 'email_reminder', 'note', 'promise_to_pay', 'escalation', 'snooze');

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    segment customer_segment NOT NULL DEFAULT 'standard',
    payment_terms_days INT NOT NULL DEFAULT 30,
    monthly_mrr DECIMAL(12,2) NOT NULL DEFAULT 0,
    contact_email VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    due_date DATE NOT NULL,
    paid_at TIMESTAMPTZ,
    status invoice_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(invoice_number)
);

CREATE TABLE collection_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    action_type action_type NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_by VARCHAR(100) NOT NULL DEFAULT 'finanzas',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_status_due ON invoices(status, due_date);
CREATE INDEX idx_actions_customer ON collection_actions(customer_id);
