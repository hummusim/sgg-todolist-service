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

func TestGetTask(t *testing.T) {
	tests := []struct {
		name   string
		input  func() (*model.Task, error)
		expect error
	}{
		{
			name: "GetTaskById_Success",
			input: func() (*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

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
						"id":         task.Id,
					}).
					Limit(limitOne).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"value",
						"completed",
						"due_date",
						"created_at",
						"updated_at",
						"deleted_at",
					},
				).AddRow(
					task.Id,
					task.Value,
					task.Completed,
					task.DueDate,
					task.CreatedAt,
					task.UpdatedAt,
					task.DeletedAt,
				))

				mock.ExpectCommit()

				svc := NewTaskRepository(db)

				return svc.GetTask(context.Background(), task.Id)
			},
			expect: nil,
		},
		{
			name: "GetTaskById_Success",
			input: func() (*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

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
						"id":         task.Id,
					}).
					Limit(limitOne).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"value",
						"completed",
						"due_date",
						"created_at",
						"updated_at",
						"deleted_at",
					},
				).AddRow(
					task.Id,
					task.Value,
					task.Completed,
					task.DueDate,
					task.CreatedAt,
					task.UpdatedAt,
					task.DeletedAt,
				))

				svc := NewTaskRepository(db)

				return svc.GetTask(context.Background(), task.Id)
			},
			expect: nil,
		},
		{
			name: "CreateLabel_ErrNoRows",
			input: func() (*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

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
						"id":         task.Id,
					}).
					Limit(limitOne).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0]).WillReturnError(sql.ErrNoRows)

				svc := NewTaskRepository(db)
				return svc.GetTask(context.Background(), task.Id)
			},
			expect: ErrTaskNotFound,
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

func TestGetTasks(t *testing.T) {
	tests := []struct {
		name   string
		input  func() ([]*model.Task, error)
		expect error
	}{
		{
			name: "GetTaskById_Success",
			input: func() ([]*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				query, _, err := psql.
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
					Offset(GetOffset(1, limitPage)).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"value",
						"completed",
						"due_date",
						"created_at",
						"updated_at",
						"deleted_at",
					},
				).AddRow(
					task.Id,
					task.Value,
					task.Completed,
					task.DueDate,
					task.CreatedAt,
					task.UpdatedAt,
					task.DeletedAt,
				))

				mock.ExpectCommit()

				svc := NewTaskRepository(db)

				return svc.GetTasks(context.Background(), 1)
			},
			expect: nil,
		},
		{
			name: "GetTaskById_Success",
			input: func() ([]*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				query, _, err := psql.
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
					Offset(GetOffset(1, limitPage)).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"value",
						"completed",
						"due_date",
						"created_at",
						"updated_at",
						"deleted_at",
					},
				).AddRow(
					task.Id,
					task.Value,
					task.Completed,
					task.DueDate,
					task.CreatedAt,
					task.UpdatedAt,
					task.DeletedAt,
				))

				svc := NewTaskRepository(db)

				return svc.GetTasks(context.Background(), 1)
			},
			expect: nil,
		},
		{
			name: "CreateLabel_ErrNoRows",
			input: func() ([]*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				query, _, err := psql.
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
					Offset(GetOffset(1, limitPage)).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(sql.ErrNoRows)

				svc := NewTaskRepository(db)
				return svc.GetTasks(context.Background(), 1)
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

func TestCreateTask(t *testing.T) {
	var (
		errCannotCreateTransaction error = errors.New("cannot create transaction")
	)

	tests := []struct {
		name   string
		input  func() (*model.Task, error)
		expect error
	}{
		{
			name: "CreateTask_Success",
			input: func() (*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				query, args, err := psql.
					Insert("tasks").
					Columns("value", "due_date").
					Values(task.Value, task.DueDate).
					Suffix("RETURNING \"id\", \"value\", \"completed\", \"due_date\", \"created_at\", \"updated_at\", \"deleted_at\"").
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnRows(sqlmock.NewRows(
					[]string{
						"id",
						"value",
						"completed",
						"due_date",
						"created_at",
						"updated_at",
						"deleted_at",
					},
				).AddRow(
					task.Id,
					task.Value,
					task.Completed,
					task.DueDate,
					task.CreatedAt,
					task.UpdatedAt,
					task.DeletedAt,
				))

				mock.ExpectCommit()

				svc := NewTaskRepository(db)

				return svc.CreateTask(context.Background(), task)
			},
			expect: nil,
		},
		{
			name: "CreateTask_ErrBeginTx",
			input: func() (*model.Task, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{},
				}

				mock.ExpectBegin().WillReturnError(errCannotCreateTransaction)

				svc := NewTaskRepository(db)

				return svc.CreateTask(context.Background(), task)
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

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name   string
		input  func() error
		expect error
	}{
		{
			name: "UpdateTask_ErrConnDone",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				timeNow := sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				}

				monkey.Patch(time.Now, func() time.Time { return timeNow.Time })

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: timeNow,
					UpdatedAt: timeNow.Time,
				}

				query, args, err := psql.
					Update("tasks").
					Set("value", task.Value).
					Set("completed", task.Completed).
					Set("due_date", task.DueDate).
					Set("updated_at", task.UpdatedAt).
					Where(sq.Eq{
						"deleted_at": nil,
						"id":         task.Id,
					}).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(
						args[0],
						args[1],
						args[2],
						args[3],
						args[4],
					).
					WillReturnError(sql.ErrConnDone)

				svc := NewTaskRepository(db)
				return svc.UpdateTask(context.Background(), &task)
			},
			expect: sql.ErrConnDone,
		},
		{
			name: "UpdateTask_Success",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				timeNow := sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				}

				monkey.Patch(time.Now, func() time.Time { return timeNow.Time })

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: timeNow,
					UpdatedAt: timeNow.Time,
				}

				query, args, err := psql.
					Update("tasks").
					Set("value", task.Value).
					Set("completed", task.Completed).
					Set("due_date", task.DueDate).
					Set("updated_at", task.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"id":         task.Id,
					}).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(
						args[0],
						args[1],
						args[2],
						args[3],
						args[4],
					).
					WillReturnRows()

				svc := NewTaskRepository(db)
				return svc.UpdateTask(context.Background(), &task)
			},
			expect: nil,
		},
		{
			name: "UpdateTask_ErrTaskNotFound",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				timeNow := sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				}

				monkey.Patch(time.Now, func() time.Time { return timeNow.Time })

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: timeNow,
					UpdatedAt: timeNow.Time,
				}

				query, args, err := psql.
					Update("tasks").
					Set("value", task.Value).
					Set("completed", task.Completed).
					Set("due_date", task.DueDate).
					Set("updated_at", task.DeletedAt.Time).
					Where(sq.Eq{
						"deleted_at": nil,
						"id":         task.Id,
					}).
					ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(
						args[0],
						args[1],
						args[2],
						args[3],
						args[4],
					).
					WillReturnError(sql.ErrNoRows)

				svc := NewTaskRepository(db)
				return svc.UpdateTask(context.Background(), &task)
			},
			expect: ErrTaskNotFound,
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

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		name   string
		input  func() error
		expect error
	}{
		{
			name: "DeleteTask_ErrNoRows",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return task.DeletedAt.Time })

				query, args, err := psql.
					Update("tasks").
					Set("deleted_at", time.Now()).
					Where(sq.Eq{
						"deleted_at": nil,
						"id":         task.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnError(sql.ErrNoRows)

				svc := NewTaskRepository(db)
				return svc.DeleteTask(context.Background(), task.Id)
			},
			expect: sql.ErrNoRows,
		},
		{
			name: "DeleteTask_Success",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return task.DeletedAt.Time })

				query, args, err := psql.
					Update("tasks").
					Set("deleted_at", time.Now()).
					Where(sq.Eq{
						"deleted_at": nil,
						"id":         task.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnResult(sqlmock.NewResult(1, 1))

				svc := NewTaskRepository(db)
				return svc.DeleteTask(context.Background(), task.Id)
			},
			expect: nil,
		},
		{
			name: "DeleteTask_ErrTaskNotFound",
			input: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				defer db.Close()

				task := model.Task{
					Id:        uuid.NewV4(),
					Value:     uuid.NewV4().String(),
					CreatedAt: time.Now(),
					DeletedAt: sql.NullTime{
						Time:  time.Now(),
						Valid: true,
					},
				}

				monkey.Patch(time.Now, func() time.Time { return task.DeletedAt.Time })

				query, args, err := psql.
					Update("tasks").
					Set("deleted_at", time.Now()).
					Where(sq.Eq{
						"deleted_at": nil,
						"id":         task.Id,
					}).ToSql()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when creating a new query", err)
				}
				row := sqlmock.NewResult(0, 0)

				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(args[0], args[1]).WillReturnResult(row)

				svc := NewTaskRepository(db)
				return svc.DeleteTask(context.Background(), task.Id)
			},
			expect: ErrTaskNotFound,
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
