import { FormEvent, useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { createAction, getCustomer, type CustomerDetail as Detail } from '../api/client'

const ACTION_TYPES = [
  { value: 'call', label: 'Llamada' },
  { value: 'email_reminder', label: 'Recordatorio email' },
  { value: 'note', label: 'Nota interna' },
  { value: 'promise_to_pay', label: 'Promesa de pago' },
  { value: 'escalation', label: 'Escalación' },
  { value: 'snooze', label: 'Pausar (snooze 72h)' },
]

function formatUSD(n: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n)
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('es-CL')
}

export default function CustomerDetail() {
  const { id } = useParams<{ id: string }>()
  const [detail, setDetail] = useState<Detail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionType, setActionType] = useState('call')
  const [notes, setNotes] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      setDetail(await getCustomer(id))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Error al cargar cliente')
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!id) return
    setSubmitting(true)
    setSubmitError(null)
    try {
      await createAction(id, {
        action_type: actionType,
        notes: notes || (actionType === 'snooze' ? 'Pausa 72h' : ''),
        created_by: 'finanzas',
      })
      setNotes('')
      await load()
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'No se pudo guardar')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return <div className="loading">Cargando detalle del cliente…</div>
  }

  if (error || !detail) {
    return (
      <div className="error-box">
        <p>{error ?? 'Cliente no encontrado'}</p>
        <Link to="/" className="back-link">← Volver al dashboard</Link>
      </div>
    )
  }

  const c = detail.customer
  const q = detail.queue_context

  return (
    <div className="detail-layout">
      <Link to="/" className="back-link">← Volver a la cola</Link>

      <section className="panel">
        <h2>{c.name}</h2>
        <p className="segment">
          {c.segment} · MRR {formatUSD(c.monthly_mrr)} · Términos {c.payment_terms_days} días
        </p>
        <p>{c.contact_email}</p>
        <div className="recommendation">
          <strong>Prioridad: </strong>
          <span className={`badge ${q.priority}`}>{q.priority}</span>
          {' '}(score {q.priority_score}) — {q.recommended_action}
        </div>
        <p style={{ marginTop: '0.75rem' }}>
          Mora: <strong>{formatUSD(q.total_overdue)}</strong>
          {q.max_days_overdue > 0 && ` · ${q.max_days_overdue} días de atraso máx.`}
        </p>
      </section>

      <section className="panel">
        <h2>Registrar acción de cobranza</h2>
        <form className="action-form" onSubmit={handleSubmit}>
          <select value={actionType} onChange={(e) => setActionType(e.target.value)} required>
            {ACTION_TYPES.map((a) => (
              <option key={a.value} value={a.value}>{a.label}</option>
            ))}
          </select>
          <textarea
            rows={3}
            placeholder="Notas (obligatorio excepto snooze)"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
          />
          {submitError && <div className="error-box">{submitError}</div>}
          <button type="submit" disabled={submitting}>
            {submitting ? 'Guardando…' : 'Guardar acción'}
          </button>
        </form>
      </section>

      <section className="panel">
        <h2>Facturas</h2>
        <table>
          <thead>
            <tr>
              <th>Número</th>
              <th>Monto</th>
              <th>Vencimiento</th>
              <th>Estado</th>
            </tr>
          </thead>
          <tbody>
            {detail.invoices.map((inv) => (
              <tr key={inv.id}>
                <td>{inv.invoice_number}</td>
                <td>{formatUSD(inv.amount)}</td>
                <td>{formatDate(inv.due_date)}</td>
                <td>{inv.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>

      <section className="panel">
        <h2>Historial de acciones</h2>
        <ul className="timeline">
          {detail.actions.length === 0 ? (
            <li>Sin acciones registradas.</li>
          ) : (
            detail.actions.map((a) => (
              <li key={a.id}>
                <strong>{a.action_type}</strong> — {a.notes}
                <div className="meta">
                  {a.created_by} · {formatDate(a.created_at)}
                </div>
              </li>
            ))
          )}
        </ul>
      </section>
    </div>
  )
}
