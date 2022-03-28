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
	ErrCommentNotFound = errors.New("comment not found")
)

type CommentRepository interface {
	CreateComment(context.Context, model.Comment) (*model.Comment, error)
	GetCommentsByTaskId(context.Context, uuid.UUID) ([]*model.Comment, error)
	DeleteCommentByTaskIdAndCommentId(ctx context.Context, taskId uuid.UUID, commentId uuid.UUID) error
}

type commentRepository struct {
	db storage.DB
}

func NewCommentRepository(db storage.DB) CommentRepository {
	return &commentRepository{
		db: db,
	}
}

func (cr *commentRepository) CreateComment(ctx context.Context, newComment model.Comment) (*model.Comment, error) {
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

	query, args, err := psql.
		Insert("comments").
		Columns("task_id", "value").
		Values(newComment.TaskId, newComment.Value).
		Suffix("RETURNING \"id\", \"task_id\", \"value\", \"created_at\", \"deleted_at\"").
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

	var comment model.Comment

	if err := row.Scan(
		&comment.Id,
		&comment.TaskId,
		&comment.Value,
		&comment.CreatedAt,
		&comment.DeletedAt,
	); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (cr *commentRepository) GetCommentsByTaskId(ctx context.Context, taskId uuid.UUID) ([]*model.Comment, error) {
	query, args, err := psql.
		Select(`
			id,
			task_id,
			value,
			created_at,
			deleted_at
		`).
		From("comments").
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

	var comments []*model.Comment = []*model.Comment{}

	for rows.Next() {
		var comment model.Comment

		if err := rows.Scan(
			&comment.Id,
			&comment.TaskId,
			&comment.Value,
			&comment.CreatedAt,
			&comment.DeletedAt,
		); err != nil {
			return nil, err
		}

		comments = append(comments, &comment)
	}

	return comments, nil
}

func (cr *commentRepository) DeleteCommentByTaskIdAndCommentId(ctx context.Context, taskId uuid.UUID, commentId uuid.UUID) error {
	query, args, err := psql.
		Update("comments").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{
			"deleted_at": nil,
			"task_id":    taskId,
			"id":         commentId,
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
		return ErrCommentNotFound
	}

	return nil
}
