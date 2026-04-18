package integration

import (
	"testing"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/scheduler"
)

func TestCronNext_Daily(t *testing.T) {
	from := time.Date(2026, 4, 18, 14, 0, 0, 0, time.UTC)
	next, err := scheduler.Next("0 2 * * *", from)
	if err != nil {
		t.Fatal(err)
	}
	expected := time.Date(2026, 4, 19, 2, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestCronNext_Every15Min(t *testing.T) {
	from := time.Date(2026, 4, 18, 14, 3, 0, 0, time.UTC)
	next, err := scheduler.Next("*/15 * * * *", from)
	if err != nil {
		t.Fatal(err)
	}
	expected := time.Date(2026, 4, 18, 14, 15, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}
