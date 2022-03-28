package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"

	"github.com/overridesh/sgg-todolist-service/internal/model"
	storage "github.com/overridesh/sgg-todolist-service/pkg/storage/sql"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskRepository interface {
	GetTask(context.Context, uuid.UUID) (*model.Task, error)
	GetTasks(context.Context, int32) ([]*model.Task, error)
	CreateTask(context.Context, model.Task) (*model.Task, error)
	UpdateTask(context.Context, *model.Task) error
	DeleteTask(context.Context, uuid.UUID) error
}

type taskRepository struct {
	db storage.DB
}

func NewTaskRepository(db storage.DB) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (tk *taskRepository) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	query, args, err := psql.
		Select(`
			id,
			value,
			completed,
			due_date,
			created_at,
			updated_at,
			deleted_at
		`).
		From("tasks").
		Where(sq.Eq{
			"deleted_at": nil,
			"id":         id,
		}).
		Limit(limitOne).
		ToSql()
	if err != nil {
		return nil, err
	}

	var task model.Task

	if err := tk.db.QueryRowContext(ctx, query, args...).Scan(
		&task.Id,
		&task.Value,
		&task.Completed,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DeletedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &task, nil
}

func (tk *taskRepository) GetTasks(ctx context.Context, page int32) ([]*model.Task, error) {
	query, args, err := psql.
		Select(`
			id,
			value,
			completed,
			due_date,
			created_at,
			updated_at,
			deleted_at
		`).
		From("tasks").
		Where(sq.Eq{
			"deleted_at": nil,
		}).
		Limit(limitPage).
		Offset(GetOffset(page, limitPage)).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tk.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []*model.Task = []*model.Task{}

	for rows.Next() {
		var task model.Task

		if err := rows.Scan(
			&task.Id,
			&task.Value,
			&task.Completed,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeletedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (tk *taskRepository) CreateTask(ctx context.Context, newTask model.Task) (*model.Task, error) {
	var (
		err error
		tx  *sql.Tx
	)

	tx, err = tk.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				zap.S().Errorf("cannot do a rollback, error: %v", err)
			}
		} else {
			if err := tx.Commit(); err != nil {
				zap.S().Errorf("cannot do a commit, error: %v", err)
			}
		}
	}()

	query, args, err := psql.
		Insert("tasks").
		Columns("value", "due_date").
		Values(newTask.Value, newTask.DueDate).
		Suffix("RETURNING \"id\", \"value\", \"completed\", \"due_date\", \"created_at\", \"updated_at\", \"deleted_at\"").
		ToSql()
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(
		ctx,
		query,
		args...,
	)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var task model.Task
	if err := row.Scan(
		&task.Id,
		&task.Value,
		&task.Completed,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DeletedAt,
	); err != nil {
		return nil, err
	}

	return &task, nil
}

func (tk *taskRepository) UpdateTask(ctx context.Context, task *model.Task) error {
	query, args, err := psql.
		Update("tasks").
		Set("value", task.Value).
		Set("completed", task.Completed).
		Set("due_date", task.DueDate).
		Set("updated_at", time.Now()).
		Where(sq.Eq{
			"deleted_at": nil,
			"id":         task.Id,
		}).
		ToSql()
	if err != nil {
		return err
	}

	if err := tk.db.QueryRowContext(ctx, query, args...).Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return err
	}

	return nil
}

func (tk *taskRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	query, args, err := psql.
		Update("tasks").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{
			"deleted_at": nil,
			"id":         id,
		}).ToSql()
	if err != nil {
		return err
	}

	result, err := tk.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrTaskNotFound
	}

	return nil
}
