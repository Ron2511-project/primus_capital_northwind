import { describe, it, expect } from 'vitest'
import {
  aggregateByPriority,
  aggregateBySegment,
  aggregateByAging,
  topOverdueCustomers,
} from './dashboardAnalytics'
import type { CollectionQueueItem } from '../api/client'

// ── Fixtures ──────────────────────────────────────────────────────────────────

function makeItem(overrides: Partial<CollectionQueueItem> = {}): CollectionQueueItem {
  return {
    customer: {
      id: crypto.randomUUID(),
      name: 'Test Customer',
      segment: 'standard',
      payment_terms_days: 30,
      monthly_mrr: 3000,
      contact_email: 'test@example.com',
      is_active: true,
      created_at: new Date().toISOString(),
    },
    total_overdue: 1000,
    max_days_overdue: 15,
    open_invoice_count: 1,
    priority: 'medium',
    priority_score: 50,
    recommended_action: 'Recordatorio personalizado',
    ...overrides,
  }
}

// ── aggregateByPriority ───────────────────────────────────────────────────────

describe('aggregateByPriority', () => {
  it('counts items per priority level', () => {
    const queue = [
      makeItem({ priority: 'critical' }),
      makeItem({ priority: 'critical' }),
      makeItem({ priority: 'high' }),
      makeItem({ priority: 'low' }),
    ]
    const result = aggregateByPriority(queue)

    const critical = result.find((r) => r.priority === 'critical')
    const high = result.find((r) => r.priority === 'high')
    const low = result.find((r) => r.priority === 'low')

    expect(critical?.count).toBe(2)
    expect(high?.count).toBe(1)
    expect(low?.count).toBe(1)
  })

  it('returns empty array for empty queue', () => {
    expect(aggregateByPriority([])).toEqual([])
  })

  it('respects priority order: critical first', () => {
    const queue = [
      makeItem({ priority: 'low' }),
      makeItem({ priority: 'critical' }),
      makeItem({ priority: 'high' }),
    ]
    const result = aggregateByPriority(queue)
    expect(result[0].priority).toBe('critical')
    expect(result[1].priority).toBe('high')
    expect(result[2].priority).toBe('low')
  })

  it('assigns correct color and label', () => {
    const queue = [makeItem({ priority: 'critical' })]
    const result = aggregateByPriority(queue)
    expect(result[0].fill).toBe('#dc2626')
    expect(result[0].name).toBe('Crítica')
  })

  it('omits priorities with no items', () => {
    const queue = [makeItem({ priority: 'medium' })]
    const result = aggregateByPriority(queue)
    expect(result.every((r) => r.priority === 'medium')).toBe(true)
    expect(result.length).toBe(1)
  })
})

// ── aggregateBySegment ────────────────────────────────────────────────────────

describe('aggregateBySegment', () => {
  it('groups count and amount by segment', () => {
    const queue = [
      makeItem({ customer: { ...makeItem().customer, segment: 'enterprise' }, total_overdue: 5000 }),
      makeItem({ customer: { ...makeItem().customer, segment: 'enterprise' }, total_overdue: 3000 }),
      makeItem({ customer: { ...makeItem().customer, segment: 'startup' }, total_overdue: 1000 }),
    ]
    const result = aggregateBySegment(queue)
    const enterprise = result.find((r) => r.segment === 'enterprise')
    const startup = result.find((r) => r.segment === 'startup')

    expect(enterprise?.count).toBe(2)
    expect(enterprise?.amount).toBe(8000)
    expect(startup?.count).toBe(1)
    expect(startup?.amount).toBe(1000)
  })

  it('sorts by amount descending', () => {
    const queue = [
      makeItem({ customer: { ...makeItem().customer, segment: 'startup' }, total_overdue: 500 }),
      makeItem({ customer: { ...makeItem().customer, segment: 'enterprise' }, total_overdue: 9000 }),
    ]
    const result = aggregateBySegment(queue)
    expect(result[0].segment).toBe('enterprise')
    expect(result[1].segment).toBe('startup')
  })

  it('returns empty array for empty queue', () => {
    expect(aggregateBySegment([])).toEqual([])
  })
})

// ── aggregateByAging ──────────────────────────────────────────────────────────

describe('aggregateByAging', () => {
  it('buckets items by days overdue', () => {
    const queue = [
      makeItem({ max_days_overdue: 0, total_overdue: 0 }),   // current
      makeItem({ max_days_overdue: 15, total_overdue: 500 }), // 1-30
      makeItem({ max_days_overdue: 45, total_overdue: 800 }), // 31-60
      makeItem({ max_days_overdue: 75, total_overdue: 1200 }), // 61-90
      makeItem({ max_days_overdue: 95, total_overdue: 3000 }), // 90+
    ]
    const result = aggregateByAging(queue)
    const keys = result.map((b) => b.key)

    expect(keys).toContain('1-30')
    expect(keys).toContain('31-60')
    expect(keys).toContain('61-90')
    expect(keys).toContain('90+')
  })

  it('sums amount correctly per bucket', () => {
    const queue = [
      makeItem({ max_days_overdue: 10, total_overdue: 400 }),
      makeItem({ max_days_overdue: 20, total_overdue: 600 }),
    ]
    const result = aggregateByAging(queue)
    const bucket = result.find((b) => b.key === '1-30')
    expect(bucket?.amount).toBe(1000)
    expect(bucket?.count).toBe(2)
  })

  it('omits buckets with no items', () => {
    const queue = [makeItem({ max_days_overdue: 95, total_overdue: 500 })]
    const result = aggregateByAging(queue)
    expect(result.length).toBe(1)
    expect(result[0].key).toBe('90+')
  })

  it('returns empty array for empty queue', () => {
    const result = aggregateByAging([])
    expect(result).toEqual([])
  })
})

// ── topOverdueCustomers ───────────────────────────────────────────────────────

describe('topOverdueCustomers', () => {
  it('returns top N customers sorted by overdue amount', () => {
    const queue = [
      makeItem({ customer: { ...makeItem().customer, name: 'A' }, total_overdue: 500 }),
      makeItem({ customer: { ...makeItem().customer, name: 'B' }, total_overdue: 9000 }),
      makeItem({ customer: { ...makeItem().customer, name: 'C' }, total_overdue: 3000 }),
    ]
    const result = topOverdueCustomers(queue, 2)
    expect(result[0].name).toBe('B')
    expect(result[1].name).toBe('C')
    expect(result.length).toBe(2)
  })

  it('excludes customers with no overdue amount', () => {
    const queue = [
      makeItem({ total_overdue: 0 }),
      makeItem({ customer: { ...makeItem().customer, name: 'X' }, total_overdue: 2000 }),
    ]
    const result = topOverdueCustomers(queue)
    expect(result.every((r) => r.amount > 0)).toBe(true)
  })

  it('truncates long names with ellipsis', () => {
    const longName = 'Empresa con nombre muy muy largo SA de CV'
    const queue = [makeItem({ customer: { ...makeItem().customer, name: longName }, total_overdue: 1000 })]
    const result = topOverdueCustomers(queue)
    expect(result[0].name.length).toBeLessThanOrEqual(22)
    expect(result[0].name.endsWith('…')).toBe(true)
    expect(result[0].fullName).toBe(longName)
  })

  it('does not truncate short names', () => {
    const queue = [makeItem({ customer: { ...makeItem().customer, name: 'Acme' }, total_overdue: 100 })]
    const result = topOverdueCustomers(queue)
    expect(result[0].name).toBe('Acme')
  })

  it('returns empty array for empty queue', () => {
    expect(topOverdueCustomers([])).toEqual([])
  })

  it('defaults to limit of 8', () => {
    const queue = Array.from({ length: 15 }, (_, i) =>
      makeItem({ customer: { ...makeItem().customer, name: `Customer ${i}` }, total_overdue: 1000 + i })
    )
    const result = topOverdueCustomers(queue)
    expect(result.length).toBe(8)
  })
})
