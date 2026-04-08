package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	recurrencedomain "example.com/taskservice/internal/domain/recurrence"
)

type RecurrenceRepository struct {
	pool *pgxpool.Pool
}

func NewRecurrenceRepository(pool *pgxpool.Pool) *RecurrenceRepository {
	return &RecurrenceRepository{pool: pool}
}

// Create сохраняет новое правило периодичности
func (r *RecurrenceRepository) Create(ctx context.Context, rule *recurrencedomain.Rule) (*recurrencedomain.Rule, error) {
	const query = `
        INSERT INTO recurrence_rules (type, params, start_date, end_date, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, type, params, start_date, end_date, created_at
    `

	// Сериализуем Params в JSONB
	paramsJSON, err := json.Marshal(rule.Params)
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, query,
		string(rule.Type),
		paramsJSON,
		rule.StartDate,
		rule.EndDate,
		rule.CreatedAt,
	)

	return scanRecurrenceRule(row)
}

// GetByID получает правило по ID
func (r *RecurrenceRepository) GetByID(ctx context.Context, id int64) (*recurrencedomain.Rule, error) {
	const query = `
        SELECT id, type, params, start_date, end_date, created_at
        FROM recurrence_rules
        WHERE id = $1
    `

	row := r.pool.QueryRow(ctx, query, id)
	rule, err := scanRecurrenceRule(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, recurrencedomain.ErrNotFound
		}
		return nil, err
	}

	return rule, nil
}

// Delete удаляет правило по ID
func (r *RecurrenceRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM recurrence_rules WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return recurrencedomain.ErrNotFound
	}

	return nil
}

// List возвращает все правила
func (r *RecurrenceRepository) List(ctx context.Context) ([]recurrencedomain.Rule, error) {
	const query = `
        SELECT id, type, params, start_date, end_date, created_at
        FROM recurrence_rules
        ORDER BY id DESC
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]recurrencedomain.Rule, 0)
	for rows.Next() {
		rule, err := scanRecurrenceRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}

	return rules, rows.Err()
}

// scanRecurrenceRule сканирует строку из БД в структуру Rule
func scanRecurrenceRule(scanner interface {
	Scan(dest ...any) error
}) (*recurrencedomain.Rule, error) {
	var (
		rule       recurrencedomain.Rule
		ruleType   string
		paramsJSON []byte
	)

	if err := scanner.Scan(
		&rule.ID,
		&ruleType,
		&paramsJSON,
		&rule.StartDate,
		&rule.EndDate,
		&rule.CreatedAt,
	); err != nil {
		return nil, err
	}

	rule.Type = recurrencedomain.Type(ruleType)

	// Десериализуем JSONB обратно в Params
	if err := json.Unmarshal(paramsJSON, &rule.Params); err != nil {
		return nil, err
	}

	return &rule, nil
}
