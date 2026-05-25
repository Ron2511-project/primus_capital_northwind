const API_URL = import.meta.env.VITE_API_URL ?? ''

export type Priority = 'critical' | 'high' | 'medium' | 'low' | 'monitor'
export type Segment = 'enterprise' | 'startup' | 'standard' | 'zombie'

export interface DashboardSummary {
  total_customers: number
  overdue_customers: number
  total_overdue_amount: number
  critical_count: number
  high_priority_count: number
  overdue_rate_percent: number
  zombie_count: number
}

export interface Customer {
  id: string
  name: string
  segment: Segment
  payment_terms_days: number
  monthly_mrr: number
  contact_email: string
}

export interface CollectionQueueItem {
  customer: Customer
  total_overdue: number
  max_days_overdue: number
  priority: Priority
  priority_score: number
  recommended_action: string
  open_invoice_count: number
}

export interface CustomerDetail {
  customer: Customer
  invoices: Array<{
    id: string
    invoice_number: string
    amount: number
    due_date: string
    status: string
  }>
  actions: Array<{
    id: string
    action_type: string
    notes: string
    created_by: string
    created_at: string
  }>
  queue_context: {
    priority: Priority
    priority_score: number
    recommended_action: string
    total_overdue: number
    max_days_overdue: number
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error ?? `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

export function getSummary() {
  return request<DashboardSummary>('/api/v1/dashboard/summary')
}

export function getQueue(segment?: string, priority?: string) {
  const params = new URLSearchParams()
  if (segment) params.set('segment', segment)
  if (priority) params.set('priority', priority)
  const q = params.toString()
  return request<CollectionQueueItem[]>(`/api/v1/collection-queue${q ? `?${q}` : ''}`)
}

export function getCustomer(id: string) {
  return request<CustomerDetail>(`/api/v1/customers/${id}`)
}

export function createAction(customerId: string, body: { action_type: string; notes: string; created_by?: string }) {
  return request(`/api/v1/customers/${customerId}/actions`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}
