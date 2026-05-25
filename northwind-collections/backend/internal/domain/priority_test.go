package domain

import "testing"

func TestCalculatePriority_EnterpriseWithinTerms(t *testing.T) {
	level, score, rec := CalculatePriority(SegmentEnterprise, 75, 45, 12000, 12000, false, nil)
	if level != PriorityMonitor {
		t.Fatalf("expected monitor, got %s (score %d)", level, score)
	}
	if rec == "" {
		t.Fatal("expected recommendation")
	}
}

func TestCalculatePriority_ZombieCritical(t *testing.T) {
	level, score, _ := CalculatePriority(SegmentZombie, 30, 120, 9000, 4500, true, nil)
	if level != PriorityCritical {
		t.Fatalf("expected critical, got %s (score %d)", level, score)
	}
	if score < 85 {
		t.Fatalf("expected high score, got %d", score)
	}
}

func TestCalculatePriority_StartupHigh(t *testing.T) {
	level, _, _ := CalculatePriority(SegmentStartup, 30, 35, 2400, 2400, false, nil)
	if level != PriorityHigh && level != PriorityCritical {
		t.Fatalf("expected high or critical for startup 35d overdue, got %s", level)
	}
}
