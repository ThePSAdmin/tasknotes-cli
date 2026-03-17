package cmd

import (
	"testing"
)

func TestComputeNextRecurrence_Weekly(t *testing.T) {
	next, err := computeNextRecurrence("FREQ=WEEKLY;BYDAY=MO", "2026-03-23", "2026-03-23")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "2026-03-30" {
		t.Errorf("expected 2026-03-30, got %q", next)
	}
}

func TestComputeNextRecurrence_Daily(t *testing.T) {
	next, err := computeNextRecurrence("FREQ=DAILY", "2026-03-17", "2026-03-17")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "2026-03-18" {
		t.Errorf("expected 2026-03-18, got %q", next)
	}
}

func TestComputeNextRecurrence_Monthly(t *testing.T) {
	next, err := computeNextRecurrence("FREQ=MONTHLY;BYMONTHDAY=15", "2026-03-15", "2026-03-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next != "2026-04-15" {
		t.Errorf("expected 2026-04-15, got %q", next)
	}
}
