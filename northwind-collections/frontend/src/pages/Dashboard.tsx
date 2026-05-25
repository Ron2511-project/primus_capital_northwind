import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  getQueue,
  getSummary,
  type CollectionQueueItem,
  type DashboardSummary,
} from '../api/client'
import DashboardCharts from '../components/DashboardCharts'

const SEGMENTS: { value: string; label: string }[] = [
  { value: '', label: 'Todos los segmentos' },
  { value: 'enterprise', label: 'Enterprise' },
  { value: 'startup', label: 'Startup' },
  { value: 'standard', label: 'Standard' },
  { value: 'zombie', label: 'Zombie' },
]

const PRIORITIES: { value: string; label: string }[] = [
  { value: '', label: 'Todas las prioridades' },
  { value: 'critical', label: 'Crítica' },
  { value: 'high', label: 'Alta' },
  { value: 'medium', label: 'Media' },
  { value: 'monitor', label: 'Monitoreo' },
]

function formatUSD(n: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n)
}

export default function Dashboard() {
  const navigate = useNavigate()
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [queue, setQueue] = useState<CollectionQueueItem[]>([])
  const [segment, setSegment] = useState('')
  const [priority, setPriority] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [s, q] = await Promise.all([
        getSummary(),
        getQueue(segment || undefined, priority || undefined),
      ])
      setSummary(s)
      setQueue(q.filter((item) => item.priority !== 'low' || item.total_overdue > 0))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Error al cargar datos')
    } finally {
      setLoading(false)
    }
  }, [segment, priority])

  useEffect(() => {
    load()
  }, [load])

  if (loading && !summary) {
    return <div className="loading">Cargando dashboard de cobranza…</div>
  }

  if (error && !summary) {
    return (
      <div className="error-box">
        <p>{error}</p>
        <button type="button" onClick={load} style={{ marginTop: '1rem' }}>
          Reintentar
        </button>
      </div>
    )
  }

  return (
    <>
      {summary && (
        <section className="metrics">
          <div className="metric-card critical">
            <div className="label">Prioridad crítica</div>
            <div className="value">{summary.critical_count}</div>
          </div>
          <div className="metric-card">
            <div className="label">Mora total</div>
            <div className="value">{formatUSD(summary.total_overdue_amount)}</div>
          </div>
          <div className="metric-card">
            <div className="label">Tasa mora (aprox.)</div>
            <div className="value">{summary.overdue_rate_percent.toFixed(1)}%</div>
          </div>
          <div className="metric-card">
            <div className="label">Clientes zombie</div>
            <div className="value">{summary.zombie_count}</div>
          </div>
          <div className="metric-card">
            <div className="label">En cola activa</div>
            <div className="value">{summary.overdue_customers}</div>
          </div>
          <div className="metric-card high">
            <div className="label">Prioridad alta</div>
            <div className="value">{summary.high_priority_count}</div>
          </div>
          <div className="metric-card">
            <div className="label">Clientes activos</div>
            <div className="value">{summary.total_customers}</div>
          </div>
        </section>
      )}

      <div className="filters">
        <select value={segment} onChange={(e) => setSegment(e.target.value)} aria-label="Segmento">
          {SEGMENTS.map((s) => (
            <option key={s.value} value={s.value}>{s.label}</option>
          ))}
        </select>
        <select value={priority} onChange={(e) => setPriority(e.target.value)} aria-label="Prioridad">
          {PRIORITIES.map((p) => (
            <option key={p.value} value={p.value}>{p.label}</option>
          ))}
        </select>
        <button type="button" className="primary" onClick={load} disabled={loading}>
          {loading ? 'Actualizando…' : 'Actualizar'}
        </button>
      </div>

      {error && <div className="error-box" style={{ marginBottom: '1rem' }}>{error}</div>}

      {!loading && <DashboardCharts queue={queue} />}

      <h2 className="section-title queue-title">Cola priorizada</h2>
      <div className="queue-table">
        <table>
          <thead>
            <tr>
              <th>Cliente</th>
              <th>Segmento</th>
              <th>Prioridad</th>
              <th>Mora</th>
              <th>Días</th>
              <th>Recomendación</th>
            </tr>
          </thead>
          <tbody>
            {queue.length === 0 ? (
              <tr>
                <td colSpan={6} style={{ textAlign: 'center', padding: '2rem' }}>
                  No hay cuentas con los filtros seleccionados.
                </td>
              </tr>
            ) : (
              queue.map((item) => (
                <tr
                  key={item.customer.id}
                  onClick={() => navigate(`/customers/${item.customer.id}`)}
                >
                  <td>
                    <strong>{item.customer.name}</strong>
                    <div className="segment">{item.customer.contact_email}</div>
                  </td>
                  <td><span className="segment">{item.customer.segment}</span></td>
                  <td>
                    <span className={`badge ${item.priority}`}>{item.priority}</span>
                    <div className="segment">score {item.priority_score}</div>
                  </td>
                  <td>{formatUSD(item.total_overdue)}</td>
                  <td>{item.max_days_overdue > 0 ? `${item.max_days_overdue}d` : '—'}</td>
                  <td style={{ maxWidth: 280 }}>{item.recommended_action}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </>
  )
}
