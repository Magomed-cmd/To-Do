package entities

import "time"

type DailyTaskMetrics struct {
	UserID         int64
	Date           time.Time
	CreatedTasks   int32
	CompletedTasks int32
	TotalTasks     int32
	UpdatedAt      time.Time
}
