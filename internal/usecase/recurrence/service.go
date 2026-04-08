package recurrence

import (
	"context"
	"fmt"
	"strings"
	"time"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo     Repository
	taskRepo TaskRepository
	now      func() time.Time
}

func NewService(repo Repository, taskRepo TaskRepository) *Service {
	return &Service{
		repo:     repo,
		taskRepo: taskRepo,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// Create создает правило и генерирует задачи
func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	// Валидация
	input.TaskTitle = strings.TrimSpace(input.TaskTitle)
	if input.TaskTitle == "" {
		return nil, fmt.Errorf("%w: task title is required", ErrInvalidInput)
	}

	if input.TaskStatus == "" {
		input.TaskStatus = taskdomain.StatusNew
	}
	if !input.TaskStatus.Valid() {
		return nil, fmt.Errorf("%w: invalid task status", ErrInvalidInput)
	}

	if !input.Type.Valid() {
		return nil, fmt.Errorf("%w: unknown recurrence type %q", ErrInvalidInput, input.Type)
	}

	// Парсим даты
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid start_date format", ErrInvalidInput)
	}

	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid end_date format", ErrInvalidInput)
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("%w: end_date must be after start_date", ErrInvalidInput)
	}

	// Создаем правило
	rule := &recurrencedomain.Rule{
		Type:      input.Type,
		Params:    input.Params,
		StartDate: startDate,
		EndDate:   endDate,
		CreatedAt: s.now(),
	}

	// Генерируем даты
	dates, err := rule.GenerateDates()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if len(dates) == 0 {
		return nil, fmt.Errorf("%w: no dates match the recurrence rule", ErrInvalidInput)
	}

	// Сохраняем правило
	createdRule, err := s.repo.Create(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("create recurrence rule: %w", err)
	}

	// Создаем задачи
	tasks := make([]taskdomain.Task, 0, len(dates))
	now := s.now()

	for _, d := range dates {
		scheduledDate := d
		task := &taskdomain.Task{
			Title:         input.TaskTitle,
			Description:   input.TaskDescription,
			Status:        input.TaskStatus,
			RecurrenceID:  &createdRule.ID,
			ScheduledDate: &scheduledDate,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		createdTask, err := s.taskRepo.Create(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("create task for date %s: %w", d.Format("2006-01-02"), err)
		}

		tasks = append(tasks, *createdTask)
	}

	return &CreateOutput{
		Rule:  createdRule,
		Tasks: tasks,
	}, nil
}

// GetByID возвращает правило по ID
func (s *Service) GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

// Delete удаляет правило и все связанные задачи
func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if err := s.taskRepo.DeleteByRecurrenceID(ctx, id); err != nil {
		return fmt.Errorf("delete tasks for recurrence %d: %w", id, err)
	}

	return s.repo.Delete(ctx, id)
}

// List возвращает все правила
func (s *Service) List(ctx context.Context) ([]recurrencedomain.Rule, error) {
	return s.repo.List(ctx)
}
