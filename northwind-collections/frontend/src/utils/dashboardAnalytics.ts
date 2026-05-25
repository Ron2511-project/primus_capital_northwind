import type { CollectionQueueItem, Priority, Segment } from '../api/client'

export const PRIORITY_LABELS: Record<Priority, string> = {
  critical: 'Crítica',
  high: 'Alta',
  medium: 'Media',
  monitor: 'Monitoreo',
  low: 'Baja',
}

export const SEGMENT_LABELS: Record<Segment, string> = {
  enterprise: 'Enterprise',
  startup: 'Startup',
  standard: 'Standard',
  zombie: 'Zombie',
}

export const PRIORITY_COLORS: Record<Priority, string> = {
  critical: '#dc2626',
  high: '#ea580c',
  medium: '#ca8a04',
  monitor: '#2563eb',
  low: '#94a3b8',
}

export const SEGMENT_COLORS: Record<Segment, string> = {
  enterprise: '#1e3a5f',
  startup: '#2d5a87',
  standard: '#64748b',
  zombie: '#7c3aed',
}

const AGING_BUCKETS = [
  { key: 'current', label: 'Al día', min: 0, max: 0 },
  { key: '1-30', label: '1–30 días', min: 1, max: 30 },
  { key: '31-60', label: '31–60 días', min: 31, max: 60 },
  { key: '61-90', label: '61–90 días', min: 61, max: 90 },
  { key: '90+', label: '+90 días', min: 91, max: Infinity },
] as const

const PRIORITY_ORDER: Priority[] = ['critical', 'high', 'medium', 'monitor', 'low']

export function aggregateByPriority(queue: CollectionQueueItem[]) {
  const counts = new Map<Priority, number>()
  for (const item of queue) {
    counts.set(item.priority, (counts.get(item.priority) ?? 0) + 1)
  }
  return PRIORITY_ORDER.filter((p) => counts.has(p)).map((priority) => ({
    priority,
    name: PRIORITY_LABELS[priority],
    count: counts.get(priority) ?? 0,
    fill: PRIORITY_COLORS[priority],
  }))
}

export function aggregateBySegment(queue: CollectionQueueItem[]) {
  const bySegment = new Map<Segment, { count: number; amount: number }>()
  for (const item of queue) {
    const seg = item.customer.segment
    const prev = bySegment.get(seg) ?? { count: 0, amount: 0 }
    bySegment.set(seg, {
      count: prev.count + 1,
      amount: prev.amount + item.total_overdue,
    })
  }
  return Array.from(bySegment.entries())
    .sort((a, b) => b[1].amount - a[1].amount)
    .map(([segment, data]) => ({
      segment,
      name: SEGMENT_LABELS[segment],
      count: data.count,
      amount: data.amount,
      fill: SEGMENT_COLORS[segment],
    }))
}

export function aggregateByAging(queue: CollectionQueueItem[]) {
  return AGING_BUCKETS.map((bucket) => {
    const items = queue.filter((item) => {
      const days = item.max_days_overdue
      return days >= bucket.min && days <= bucket.max
    })
    const amount = items.reduce((sum, i) => sum + i.total_overdue, 0)
    return {
      key: bucket.key,
      name: bucket.label,
      count: items.length,
      amount,
    }
  }).filter((b) => b.count > 0)
}

export function topOverdueCustomers(queue: CollectionQueueItem[], limit = 8) {
  return [...queue]
    .filter((item) => item.total_overdue > 0)
    .sort((a, b) => b.total_overdue - a.total_overdue)
    .slice(0, limit)
    .map((item) => ({
      id: item.customer.id,
      name:
        item.customer.name.length > 22
          ? `${item.customer.name.slice(0, 20)}…`
          : item.customer.name,
      fullName: item.customer.name,
      amount: item.total_overdue,
      days: item.max_days_overdue,
      priority: item.priority,
      fill: PRIORITY_COLORS[item.priority],
    }))
}
