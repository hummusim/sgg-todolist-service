package model

import (
	"database/sql"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Label struct {
	Id        uuid.UUID
	TaskId    uuid.UUID
	Value     string
	CreatedAt time.Time
	DeletedAt sql.NullTime
}
