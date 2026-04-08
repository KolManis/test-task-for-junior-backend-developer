package handlers

import (
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
}

type taskDTO struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Status        taskdomain.Status `json:"status"`
	RecurrenceID  *int64            `json:"recurrence_id,omitempty"`
	ScheduledDate *string           `json:"scheduled_date,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	dto := taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		RecurrenceID: task.RecurrenceID,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}

	if task.ScheduledDate != nil {
		s := task.ScheduledDate.Format("2006-01-02")
		dto.ScheduledDate = &s
	}

	return dto
}

type createRecurrenceRequestDTO struct {
	Task struct {
		Title       string            `json:"title"`
		Description string            `json:"description"`
		Status      taskdomain.Status `json:"status"`
	} `json:"task"`

	Type      recurrencedomain.Type   `json:"type"`
	Params    recurrencedomain.Params `json:"params"`
	StartDate string                  `json:"start_date"`
	EndDate   string                  `json:"end_date"`
}

type recurrenceRuleDTO struct {
	ID        int64                   `json:"id"`
	Type      recurrencedomain.Type   `json:"type"`
	Params    recurrencedomain.Params `json:"params"`
	StartDate string                  `json:"start_date"`
	EndDate   string                  `json:"end_date"`
	CreatedAt time.Time               `json:"created_at"`
}

func newRecurrenceRuleDTO(rule *recurrencedomain.Rule) recurrenceRuleDTO {
	return recurrenceRuleDTO{
		ID:        rule.ID,
		Type:      rule.Type,
		Params:    rule.Params,
		StartDate: rule.StartDate.Format("2006-01-02"),
		EndDate:   rule.EndDate.Format("2006-01-02"),
		CreatedAt: rule.CreatedAt,
	}
}

type createRecurrenceResponseDTO struct {
	Rule  recurrenceRuleDTO `json:"rule"`
	Tasks []taskDTO         `json:"tasks"`
}
