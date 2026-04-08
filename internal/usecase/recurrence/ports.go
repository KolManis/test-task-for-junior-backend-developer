package recurrence

import (
	"context"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

// Repository для правил периодичности
type Repository interface {
	Create(ctx context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error)
	GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]recurrencedomain.Rule, error)
}

// TaskRepository (часть интерфейса, нужная для работы с задачами)
type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	DeleteByRecurrenceID(ctx context.Context, recurrenceID int64) error
}

// Usecase интерфейс
type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*CreateOutput, error)
	GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]recurrencedomain.Rule, error)
}

// CreateInput входные данные
type CreateInput struct {
	TaskTitle       string
	TaskDescription string
	TaskStatus      taskdomain.Status
	Type            recurrencedomain.Type
	Params          recurrencedomain.Params
	StartDate       string
	EndDate         string
}

// CreateOutput результат создания
type CreateOutput struct {
	Rule  *recurrencedomain.Rule
	Tasks []taskdomain.Task
}
