package domain

type DashboardSummary struct {
	TotalCustomers      int     `json:"total_customers"`
	OverdueCustomers    int     `json:"overdue_customers"`
	TotalOverdueAmount  float64 `json:"total_overdue_amount"`
	CriticalCount       int     `json:"critical_count"`
	HighPriorityCount   int     `json:"high_priority_count"`
	OverdueRatePercent  float64 `json:"overdue_rate_percent"`
	ZombieCount         int     `json:"zombie_count"`
}
