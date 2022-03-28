package model

import (
	"database/sql"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Task struct {
	Id        uuid.UUID
	Value     string
	Completed bool
	DueDate   sql.NullTime
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
	Comments  []*Comment
	Labels    []*Label
}
