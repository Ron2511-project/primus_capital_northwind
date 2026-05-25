import { useMemo } from 'react'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import type { CollectionQueueItem } from '../api/client'
import {
  aggregateByAging,
  aggregateByPriority,
  aggregateBySegment,
  topOverdueCustomers,
} from '../utils/dashboardAnalytics'

function formatUSD(n: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(n)
}

function formatUSDCompact(n: number) {
  if (n >= 1_000_000) return `$${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `$${(n / 1_000).toFixed(0)}k`
  return formatUSD(n)
}

interface Props {
  queue: CollectionQueueItem[]
}

export default function DashboardCharts({ queue }: Props) {
  const byPriority = useMemo(() => aggregateByPriority(queue), [queue])
  const bySegment = useMemo(() => aggregateBySegment(queue), [queue])
  const byAging = useMemo(() => aggregateByAging(queue), [queue])
  const topDebtors = useMemo(() => topOverdueCustomers(queue), [queue])

  const hasData = queue.length > 0

  if (!hasData) {
    return (
      <section className="charts-section">
        <h2 className="section-title">Análisis visual</h2>
        <div className="charts-empty">No hay datos para graficar con los filtros actuales.</div>
      </section>
    )
  }

  return (
    <section className="charts-section">
      <h2 className="section-title">Análisis visual</h2>
      <div className="charts-grid">
        <div className="chart-panel">
          <h3>Distribución por prioridad</h3>
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie
                data={byPriority}
                dataKey="count"
                nameKey="name"
                cx="50%"
                cy="50%"
                innerRadius={55}
                outerRadius={90}
                paddingAngle={2}
                label={({ name, percent }) =>
                  percent > 0.05 ? `${name} ${(percent * 100).toFixed(0)}%` : ''
                }
              >
                {byPriority.map((entry) => (
                  <Cell key={entry.priority} fill={entry.fill} />
                ))}
              </Pie>
              <Tooltip formatter={(value: number) => [`${value} clientes`, 'Cantidad']} />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>

        <div className="chart-panel">
          <h3>Mora por segmento</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={bySegment} margin={{ top: 8, right: 8, left: 8, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="name" tick={{ fontSize: 11 }} />
              <YAxis tickFormatter={formatUSDCompact} tick={{ fontSize: 11 }} width={52} />
              <Tooltip
                formatter={(value: number, name: string) => {
                  if (name === 'amount') return [formatUSD(value), 'Mora total']
                  return [value, 'Clientes']
                }}
              />
              <Legend />
              <Bar dataKey="amount" name="Mora total" fill="#1e3a5f" radius={[4, 4, 0, 0]} />
              <Bar dataKey="count" name="Clientes" fill="#94a3b8" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="chart-panel">
          <h3>Antigüedad de mora</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={byAging} margin={{ top: 8, right: 8, left: 8, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis dataKey="name" tick={{ fontSize: 11 }} />
              <YAxis yAxisId="left" tick={{ fontSize: 11 }} allowDecimals={false} />
              <YAxis
                yAxisId="right"
                orientation="right"
                tickFormatter={formatUSDCompact}
                tick={{ fontSize: 11 }}
                width={52}
              />
              <Tooltip
                formatter={(value: number, name: string) => {
                  if (name === 'amount') return [formatUSD(value), 'Mora en tramo']
                  return [value, 'Clientes']
                }}
              />
              <Legend />
              <Bar yAxisId="left" dataKey="count" name="Clientes" fill="#2d5a87" radius={[4, 4, 0, 0]} />
              <Bar yAxisId="right" dataKey="amount" name="Mora" fill="#dc2626" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="chart-panel chart-panel-wide">
          <h3>Top morosos por monto</h3>
          <ResponsiveContainer width="100%" height={Math.max(220, topDebtors.length * 36)}>
            <BarChart
              data={topDebtors}
              layout="vertical"
              margin={{ top: 4, right: 16, left: 4, bottom: 4 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" horizontal={false} />
              <XAxis type="number" tickFormatter={formatUSDCompact} tick={{ fontSize: 11 }} />
              <YAxis type="category" dataKey="name" width={120} tick={{ fontSize: 11 }} />
              <Tooltip
                formatter={(value: number) => [formatUSD(value), 'Mora']}
                labelFormatter={(_, payload) => {
                  const row = payload?.[0]?.payload as { fullName?: string; days?: number } | undefined
                  if (!row) return ''
                  return `${row.fullName} · ${row.days}d atraso`
                }}
              />
              <Bar dataKey="amount" name="Mora" radius={[0, 4, 4, 0]}>
                {topDebtors.map((entry) => (
                  <Cell key={entry.id} fill={entry.fill} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </section>
  )
}
