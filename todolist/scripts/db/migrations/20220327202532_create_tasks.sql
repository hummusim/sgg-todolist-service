-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks (
    id UUID DEFAULT gen_random_uuid(),
    value VARCHAR(255),
    completed BOOLEAN DEFAULT FALSE,
    due_date TIMESTAMPTZ DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	deleted_at TIMESTAMPTZ,
    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tasks
-- +goose StatementEnd