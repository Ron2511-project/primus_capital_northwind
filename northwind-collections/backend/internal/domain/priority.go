package domain

import "time"

// CalculatePriority determina foco de cobranza según segmento, mora y monto.
// Supuesto documentado: enterprise tolera hasta payment_terms_days; zombie = 90+ días sin pago.
func CalculatePriority(
	segment Segment,
	paymentTermsDays int,
	maxDaysOverdue int,
	totalOverdue float64,
	monthlyMRR float64,
	hasUnpaid90Days bool,
	lastAction *CollectionAction,
) (PriorityLevel, int, string) {
	score := 0
	recommended := "Revisar cuenta"

	switch {
	case segment == SegmentZombie || hasUnpaid90Days:
		score = 100
		recommended = "Llamada urgente + evaluar suspensión de servicio"
	case maxDaysOverdue <= 0:
		if segment == SegmentStartup && totalOverdue > 0 {
			score = 25
			recommended = "Monitorear — posible problema de caja"
			return PriorityMonitor, score, recommended
		}
		return PriorityLow, 0, "Sin mora — no contactar"

	case segment == SegmentEnterprise:
		// Solo escalar si supera sus términos contractuales (ej. 75 días)
		if maxDaysOverdue < paymentTermsDays {
			score = 15
			return PriorityMonitor, score, "Dentro de ciclo de pago enterprise — no enviar recordatorio genérico"
		}
		score = 55 + min(maxDaysOverdue-paymentTermsDays, 30)
		recommended = "Contacto personalizado (no email masivo)"

	case segment == SegmentStartup:
		score = 70 + min(maxDaysOverdue*2, 25)
		if totalOverdue > monthlyMRR {
			score += 15
		}
		recommended = "Llamada empática — verificar situación de caja"

	default:
		score = 40 + min(maxDaysOverdue*3, 40)
		recommended = "Recordatorio personalizado"
	}

	// Peso por monto en mora (ticket hasta 15k)
	if totalOverdue >= 10000 {
		score += 20
	} else if totalOverdue >= 3000 {
		score += 10
	}

	// Reducir ruido si hubo acción reciente
	if lastAction != nil && time.Since(lastAction.CreatedAt) < 72*time.Hour {
		if lastAction.ActionType == ActionSnooze {
			return PriorityLow, score / 2, "En pausa — respetar snooze"
		}
		score = max(score-20, 0)
	}

	level := PriorityMedium
	switch {
	case score >= 85:
		level = PriorityCritical
	case score >= 65:
		level = PriorityHigh
	case score >= 40:
		level = PriorityMedium
	case score >= 20:
		level = PriorityMonitor
	default:
		level = PriorityLow
	}

	return level, score, recommended
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
