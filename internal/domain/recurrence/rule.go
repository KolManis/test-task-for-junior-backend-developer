package recurrence

import "time"

type Type string

const (
	TypeDaily         Type = "daily"          // каждый N-й день
	TypeMonthlyDay    Type = "monthly_day"    // определенное число месяца
	TypeSpecificDates Type = "specific_dates" // конкретные даты
	TypeEvenOddDays   Type = "even_odd_days"  // четные/нечетные дни
)

type Params struct {
	EveryNDays *int     `json:"every_n_days,omitempty"` // для daily
	DayOfMonth *int     `json:"day_of_month,omitempty"` // для monthly_day (1-30)
	Dates      []string `json:"dates,omitempty"`        // для specific_dates
	Parity     *string  `json:"parity,omitempty"`       // для even_odd_days ("even"/"odd")
}

type Rule struct {
	ID        int64
	Type      Type
	Params    Params
	StartDate time.Time
	EndDate   time.Time
	CreatedAt time.Time
}
