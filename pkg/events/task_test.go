package events

import (
	"testing"
	"time"
)

func TestTaskEvent_IsZero(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		event  TaskEvent
		expect bool
	}{
		{name: "empty", event: TaskEvent{}, expect: true},
		{name: "missing type", event: TaskEvent{ID: "id"}, expect: true},
		{name: "missing id", event: TaskEvent{Type: TaskEventCreated}, expect: true},
		{name: "valid", event: TaskEvent{ID: "id", Type: TaskEventCompleted, CreatedAt: now}, expect: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.IsZero(); got != tt.expect {
				t.Fatalf("expected %v, got %v", tt.expect, got)
			}
		})
	}
}
