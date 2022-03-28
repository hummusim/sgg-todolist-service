package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	uuid "github.com/satori/go.uuid"

	"github.com/overridesh/sgg-todolist-service/internal/model"
)

func TestCreateComment(t *testing.T) {
	var (
		errCannotCreateTransaction error = errors.New("cannot create transaction")
	)

	tests := []struct {
		name   string
		input  func() (*model.Comment, error)
		expect error
	}{
		{
			name: "CreateComment_Success",
			input: func() (*model.Comment, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				query, args, err := psql.
					Insert("comments").
					Columns("task_id", "value").
					Values(newComment.TaskId, newComment.Value).
					Suffix("RETURNING \"id\", \"task_id\", \"value\", \"created_at\", \"deleted_at\"").
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"task_id",
						"value",
						"created_at",
						"deleted_at",
					},
				).AddRow(
					newComment.Id,
					newComment.TaskId,
					newComment.Value,
					newComment.CreatedAt,
					newComment.DeletedAt,
				))

				mock.ExpectCommit()

				svc := NewCommentRepository(db)

				return svc.CreateComment(context.Background(), model.Comment{
					TaskId: newComment.TaskId,
					Value:  newComment.Value,
				})
			},
			expect: nil,
		},
		{
			name: "CreateComment_ErrBeginTx",
			input: func() (*model.Comment, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				mock.ExpectBegin().WillReturnError(errCannotCreateTransaction)
				svc := NewCommentRepository(db)

				return svc.CreateComment(context.Background(), model.Comment{})
			},
			expect: errCannotCreateTransaction,
		},
		{
			name: "CreateComment_ErrRow",
			input: func() (*model.Comment, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				query, args, err := psql.
					Insert("comments").
					Columns("task_id", "value").
					Values(newComment.TaskId, newComment.Value).
					Suffix("RETURNING \"id\", \"task_id\", \"value\", \"created_at\", \"deleted_at\"").
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectBegin()

				mock.ExpectQuery(
					regexp.QuoteMeta(query)).
					WithArgs(args[0], args[1]).
					WillReturnError(ErrCommentNotFound)

				mock.ExpectRollback()

				svc := NewCommentRepository(db)
				return svc.CreateComment(context.Background(), newComment)
			},
			expect: ErrCommentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input()
			if err != tt.expect {
				t.Errorf("expect values are equals, but got diferent, output: %v, expect: %v", err, tt.expect)
			}
		})
	}
}

func TestGetCommentsByTaskId(t *testing.T) {
	tests := []struct {
		name   string
		input  func() ([]*model.Comment, error)
		expect error
	}{
		{
			name: "GetCommentByTaskId_Success",
			input: func() ([]*model.Comment, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

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
						"task_id":    newComment.TaskId,
					}).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"task_id",
						"value",
						"created_at",
						"deleted_at",
					},
				).AddRow(
					newComment.Id,
					newComment.TaskId,
					newComment.Value,
					newComment.CreatedAt,
					newComment.DeletedAt,
				))

				mock.ExpectCommit()

				svc := NewCommentRepository(db)

				return svc.GetCommentsByTaskId(context.Background(), newComment.TaskId)
			},
			expect: nil,
		},
		{
			name: "GetCommentByTaskId_Success",
			input: func() ([]*model.Comment, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

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
						"task_id":    newComment.TaskId,
					}).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"task_id",
						"value",
						"created_at",
						"deleted_at",
					},
				).AddRow(
					newComment.Id,
					newComment.TaskId,
					newComment.Value,
					newComment.CreatedAt,
					newComment.DeletedAt,
				))

				svc := NewCommentRepository(db)

				return svc.GetCommentsByTaskId(context.Background(), newComment.TaskId)
			},
			expect: nil,
		},
		{
			name: "CreateComment_ErrNoRows",
			input: func() ([]*model.Comment, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

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
						"task_id":    newComment.TaskId,
					}).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0]).WillReturnError(sql.ErrNoRows)

				svc := NewCommentRepository(db)
				return svc.GetCommentsByTaskId(context.Background(), newComment.TaskId)
			},
			expect: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input()
			if err != tt.expect {
				t.Errorf("expect values are equals, but got diferent, output: %v, expect: %v", err, tt.expect)
			}
		})
	}
}

func TestDeleteCommentByTaskIdAndCommentId(t *testing.T) {
	tests := []struct {
		name   string
		input  func() error
		expect error
	}{
		{
			name: "DeleteComment_Success",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return newComment.DeletedAt.Time })

				query, args, err := psql.
					Update("comments").
					Set("deleted_at", newComment.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newComment.TaskId,
						"id":         newComment.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1], args[2]).WillReturnError(sql.ErrNoRows)

				svc := NewCommentRepository(db)
				return svc.DeleteCommentByTaskIdAndCommentId(context.Background(), newComment.TaskId, newComment.Id)
			},
			expect: sql.ErrNoRows,
		},
		{
			name: "DeleteComment_Success",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return newComment.DeletedAt.Time })

				query, args, err := psql.
					Update("comments").
					Set("deleted_at", newComment.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newComment.TaskId,
						"id":         newComment.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1], args[2]).WillReturnResult(sqlmock.NewResult(1, 1))

				svc := NewCommentRepository(db)
				return svc.DeleteCommentByTaskIdAndCommentId(context.Background(), newComment.TaskId, newComment.Id)
			},
			expect: nil,
		},
		{
			name: "DeleteComment_Success",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newComment := model.Comment{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return newComment.DeletedAt.Time })

				query, args, err := psql.
					Update("comments").
					Set("deleted_at", newComment.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newComment.TaskId,
						"id":         newComment.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				row := sqlmock.NewResult(0, 0)

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1], args[2]).WillReturnResult(row)

				svc := NewCommentRepository(db)
				return svc.DeleteCommentByTaskIdAndCommentId(context.Background(), newComment.TaskId, newComment.Id)
			},
			expect: ErrCommentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input()
			if err != tt.expect {
				t.Errorf("expect values are equals, but got diferent, output: %v, expect: %v", err, tt.expect)
			}
		})
	}
}
