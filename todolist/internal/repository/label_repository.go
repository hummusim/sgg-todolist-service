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
	ErrLabelNotFound      = errors.New("label not found")
	ErrLabelAlreadyExists = errors.New("label already exists")
)

type LabelRepository interface {
	CreateLabel(context.Context, model.Label) (*model.Label, error)
	GetLabelsByTaskId(context.Context, uuid.UUID) ([]*model.Label, error)
	DeleteLabelByTaskIdAndLabelId(ctx context.Context, taskId uuid.UUID, labelId uuid.UUID) error
}

type labelRepository struct {
	db storage.DB
}

func NewLabelRepository(db storage.DB) LabelRepository {
	return &labelRepository{
		db: db,
	}
}

func (cr *labelRepository) CreateLabel(ctx context.Context, newLabel model.Label) (*model.Label, error) {
	var (
		err error
		tx  *sql.Tx
	)

	tx, err = cr.db.BeginTx(ctx, nil)
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

	query, args, err := psql.Select("COUNT(id)").
		From("labels").
		Where(sq.Eq{
			"deleted_at": nil,
			"task_id":    newLabel.TaskId,
			"value":      newLabel.Value,
		}).ToSql()

	row := tx.QueryRowContext(
		ctx,
		query,
		args...,
	)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var count int64
	if err := row.Scan(
		&count,
	); err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, ErrLabelAlreadyExists
	}

	query, args, err = psql.
		Insert("labels").
		Columns("task_id", "value").
		Values(newLabel.TaskId, newLabel.Value).
		Suffix("RETURNING \"id\", \"task_id\", \"value\", \"created_at\", \"deleted_at\"").
		ToSql()
	if err != nil {
		return nil, err
	}

	row = tx.QueryRowContext(
		ctx,
		query,
		args...,
	)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var label model.Label
	if err := row.Scan(
		&label.Id,
		&label.TaskId,
		&label.Value,
		&label.CreatedAt,
		&label.DeletedAt,
	); err != nil {
		return nil, err
	}

	return &label, nil
}

func (cr *labelRepository) GetLabelsByTaskId(ctx context.Context, taskId uuid.UUID) ([]*model.Label, error) {
	query, args, err := psql.
		Select(`
			id,
			task_id,
			value,
			created_at,
			deleted_at
		`).
		From("labels").
		Where(sq.Eq{
			"deleted_at": nil,
			"task_id":    taskId,
		}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := cr.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var labels []*model.Label = []*model.Label{}

	for rows.Next() {
		var label model.Label

		if err := rows.Scan(
			&label.Id,
			&label.TaskId,
			&label.Value,
			&label.CreatedAt,
			&label.DeletedAt,
		); err != nil {
			return nil, err
		}

		labels = append(labels, &label)
	}

	return labels, nil
}

func (cr *labelRepository) DeleteLabelByTaskIdAndLabelId(ctx context.Context, taskId uuid.UUID, labelId uuid.UUID) error {
	query, args, err := psql.
		Update("labels").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{
			"deleted_at": nil,
			"task_id":    taskId,
			"id":         labelId,
		}).ToSql()
	if err != nil {
		return err
	}

	result, err := cr.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrLabelNotFound
	}

	return nil
}
