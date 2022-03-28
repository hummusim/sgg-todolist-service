package sql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const maxRetries int = 3

// DB interface
type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(ctx context.Context, sql string, arguments ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, sql string, optionsAndArgs ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, sql string, optionsAndArgs ...interface{}) *sql.Row
	Close() error
}

// NewConnection create a connection to postgres database
func NewConnection(dbConn string) (DB, error) {
	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		zap.L().Error(fmt.Sprintf("cannot create connection to db, error: %v", err))
		return nil, err
	}

	tries := maxRetries
	for tries >= 0 {
		if err := db.Ping(); err != nil {
			zap.L().Warn(fmt.Sprintf("cannot do a ping connection to db, error: %v", err))
			if err != nil {
				if tries == 0 {
					return nil, err
				}

				tries = tries - 1
				continue
			}

		}
		break
	}

	return db, nil
}
