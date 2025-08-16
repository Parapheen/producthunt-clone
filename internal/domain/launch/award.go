package launch

import (
	"time"

	"github.com/google/uuid"
)

// Award represents an award type like product_of_day/week/month
type Award struct {
	ID          string
	Code        string
	Name        string
	Description string
	Icon        string
}

// LaunchAward represents a concrete award assigned to a launch for a specific period
type LaunchAward struct {
	ID         string
	LaunchID   uuid.UUID
	Award      Award
	PeriodDate time.Time // canonical date of period (e.g., day, week's Monday, month's 1st)
	AwardedAt  time.Time
}

const (
	AwardCodeProductOfDay   = "product_of_day"
	AwardCodeProductOfWeek  = "product_of_week"
	AwardCodeProductOfMonth = "product_of_month"
	AwardCodeProductOfYear  = "product_of_year"
)
