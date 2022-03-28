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
	"github.com/overridesh/sgg-todolist-service/internal/model"
	uuid "github.com/satori/go.uuid"
)

func TestCreateLabel(t *testing.T) {
	var (
		errCannotCreateTransaction error = errors.New("cannot create transaction")
	)

	tests := []struct {
		name   string
		input  func() (*model.Label, error)
		expect error
	}{
		{
			name: "CreateLabel_ErrLabelAlreadyExists",
			input: func() (*model.Label, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				query, args, err := psql.Select("COUNT(id)").
					From("labels").
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
						"value":      newLabel.Value,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"COUNT",
					},
				).AddRow(
					1,
				))

				svc := NewLabelRepository(db)

				return svc.CreateLabel(context.Background(), newLabel)
			},
			expect: ErrLabelAlreadyExists,
		},
		{
			name: "CreateLabel_Success",
			input: func() (*model.Label, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				query, args, err := psql.Select("COUNT(id)").
					From("labels").
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
						"value":      newLabel.Value,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"COUNT",
					},
				).AddRow(
					0,
				))

				query, args, err = psql.
					Insert("labels").
					Columns("task_id", "value").
					Values(newLabel.TaskId, newLabel.Value).
					Suffix("RETURNING \"id\", \"task_id\", \"value\", \"created_at\", \"deleted_at\"").
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"task_id",
						"value",
						"created_at",
						"deleted_at",
					},
				).AddRow(
					newLabel.Id,
					newLabel.TaskId,
					newLabel.Value,
					newLabel.CreatedAt,
					newLabel.DeletedAt,
				))

				mock.ExpectCommit()

				svc := NewLabelRepository(db)

				return svc.CreateLabel(context.Background(), newLabel)
			},
			expect: nil,
		},
		{
			name: "CreateLabel_ErrBeginTx",
			input: func() (*model.Label, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				mock.ExpectBegin().WillReturnError(errCannotCreateTransaction)

				svc := NewLabelRepository(db)

				return svc.CreateLabel(context.Background(), newLabel)
			},
			expect: errCannotCreateTransaction,
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

func TestGetLabelsByTaskId(t *testing.T) {
	tests := []struct {
		name   string
		input  func() ([]*model.Label, error)
		expect error
	}{
		{
			name: "GetLabelsByTaskId_Success",
			input: func() ([]*model.Label, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
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
					From("labels").
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
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
					newLabel.Id,
					newLabel.TaskId,
					newLabel.Value,
					newLabel.CreatedAt,
					newLabel.DeletedAt,
				))

				mock.ExpectCommit()

				svc := NewLabelRepository(db)

				return svc.GetLabelsByTaskId(context.Background(), newLabel.TaskId)
			},
			expect: nil,
		},
		{
			name: "GetLabelsByTaskId_Success",
			input: func() ([]*model.Label, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
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
					From("labels").
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
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
					newLabel.Id,
					newLabel.TaskId,
					newLabel.Value,
					newLabel.CreatedAt,
					newLabel.DeletedAt,
				))

				svc := NewLabelRepository(db)

				return svc.GetLabelsByTaskId(context.Background(), newLabel.TaskId)
			},
			expect: nil,
		},
		{
			name: "CreateLabel_ErrNoRows",
			input: func() ([]*model.Label, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
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
					From("labels").
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
					}).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0]).WillReturnError(sql.ErrNoRows)

				svc := NewLabelRepository(db)
				return svc.GetLabelsByTaskId(context.Background(), newLabel.TaskId)
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

func TestDeleteLabelByTaskIdAndLabelId(t *testing.T) {
	tests := []struct {
		name   string
		input  func() error
		expect error
	}{
		{
			name: "DeleteLabel_ErrConnDone",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return newLabel.DeletedAt.Time })

				query, args, err := psql.
					Update("labels").
					Set("deleted_at", newLabel.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
						"id":         newLabel.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1], args[2]).WillReturnError(sql.ErrConnDone)

				svc := NewLabelRepository(db)
				return svc.DeleteLabelByTaskIdAndLabelId(context.Background(), newLabel.TaskId, newLabel.Id)
			},
			expect: sql.ErrConnDone,
		},
		{
			name: "DeleteLabels_Success",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return newLabel.DeletedAt.Time })

				query, args, err := psql.
					Update("labels").
					Set("deleted_at", newLabel.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
						"id":         newLabel.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1], args[2]).WillReturnResult(sqlmock.NewResult(1, 1))

				svc := NewLabelRepository(db)
				return svc.DeleteLabelByTaskIdAndLabelId(context.Background(), newLabel.TaskId, newLabel.Id)
			},
			expect: nil,
		},
		{
			name: "DeleteLabels_ErrLabelNotFound",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				newLabel := model.Label{
					Id:        uuid.NewV4(),
					TaskId:    uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return newLabel.DeletedAt.Time })

				query, args, err := psql.
					Update("labels").
					Set("deleted_at", newLabel.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"task_id":    newLabel.TaskId,
						"id":         newLabel.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				row := sqlmock.NewResult(0, 0)

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1], args[2]).WillReturnResult(row)

				svc := NewLabelRepository(db)
				return svc.DeleteLabelByTaskIdAndLabelId(context.Background(), newLabel.TaskId, newLabel.Id)
			},
			expect: ErrLabelNotFound,
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
