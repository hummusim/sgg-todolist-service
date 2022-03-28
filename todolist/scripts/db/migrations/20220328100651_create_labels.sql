-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS labels (
    id UUID DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL,
    value TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ,
    PRIMARY KEY (id),
    CONSTRAINT FK_task_id FOREIGN KEY (task_id)
    REFERENCES tasks(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE labels
-- +goose StatementEnd