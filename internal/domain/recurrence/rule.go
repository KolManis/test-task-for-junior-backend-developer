package recurrence

import (
	"fmt"
	"time"
)

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

const MaxGeneratedTasks = 500

// GenerateDates возвращает список дат, на которые нужно создать задачи
func (r *Rule) GenerateDates() ([]time.Time, error) {
	var dates []time.Time

	switch r.Type {
	case TypeDaily:
		dates = r.generateDaily()
	case TypeMonthlyDay:
		dates = r.generateMonthlyDay()
	case TypeSpecificDates:
		dates = r.generateSpecificDates()
	case TypeEvenOddDays:
		dates = r.generateEvenOddDays()
	default:
		return nil, fmt.Errorf("unknown recurrence type: %s", r.Type)
	}

	if len(dates) > MaxGeneratedTasks {
		return nil, fmt.Errorf("too many tasks would be generated: %d (max %d)", len(dates), MaxGeneratedTasks)
	}

	return dates, nil
}

// generateDaily: каждый N-й день
func (r *Rule) generateDaily() []time.Time {
	step := 1
	if r.Params.EveryNDays != nil && *r.Params.EveryNDays > 0 {
		step = *r.Params.EveryNDays
	}

	var dates []time.Time
	for d := r.StartDate; !d.After(r.EndDate); d = d.AddDate(0, 0, step) {
		dates = append(dates, d)
	}
	return dates
}

// generateMonthlyDay: каждый месяц в определенный день (1-30)
func (r *Rule) generateMonthlyDay() []time.Time {
	if r.Params.DayOfMonth == nil {
		return nil
	}
	day := *r.Params.DayOfMonth
	if day < 1 || day > 30 {
		return nil
	}

	var dates []time.Time

	// Начинаем с первого месяца, где есть этот день
	current := time.Date(r.StartDate.Year(), r.StartDate.Month(), day, 0, 0, 0, 0, time.UTC)
	if current.Before(r.StartDate) {
		current = current.AddDate(0, 1, 0)
	}

	for !current.After(r.EndDate) {
		// Проверяем, что день не "съехал" (например, 30 февраля → 2 марта)
		if current.Day() == day {
			dates = append(dates, current)
		}
		current = current.AddDate(0, 1, 0)
	}

	return dates
}

// generateSpecificDates: только конкретные даты
func (r *Rule) generateSpecificDates() []time.Time {
	var dates []time.Time
	seen := make(map[string]bool)

	for _, dateStr := range r.Params.Dates {
		if seen[dateStr] {
			continue
		}
		seen[dateStr] = true

		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if !d.Before(r.StartDate) && !d.After(r.EndDate) {
			dates = append(dates, d)
		}
	}

	return dates
}

// generateEvenOddDays: четные или нечетные дни месяца
func (r *Rule) generateEvenOddDays() []time.Time {
	if r.Params.Parity == nil {
		return nil
	}

	wantEven := *r.Params.Parity == "even"

	var dates []time.Time
	for d := r.StartDate; !d.After(r.EndDate); d = d.AddDate(0, 0, 1) {
		isEven := d.Day()%2 == 0
		if isEven == wantEven {
			dates = append(dates, d)
		}
	}
	return dates
}

func (t Type) Valid() bool {
	switch t {
	case TypeDaily, TypeMonthlyDay, TypeSpecificDates, TypeEvenOddDays:
		return true
	}
	return false
}
