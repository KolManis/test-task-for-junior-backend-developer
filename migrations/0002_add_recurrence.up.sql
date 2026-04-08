CREATE TABLE IF NOT EXISTS recurrence_rules (
    id         BIGSERIAL PRIMARY KEY,
    type       TEXT      NOT NULL,      -- daily, monthly_day, specific_dates, even_odd_days
    params     JSONB     NOT NULL,      
    start_date DATE      NOT NULL,
    end_date   DATE      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_dates CHECK (end_date >= start_date)
);

ALTER TABLE tasks 
    ADD COLUMN IF NOT EXISTS recurrence_id BIGINT REFERENCES recurrence_rules(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS scheduled_date DATE;

CREATE INDEX IF NOT EXISTS idx_tasks_recurrence_id ON tasks(recurrence_id);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_date ON tasks(scheduled_date);