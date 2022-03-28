package model

import (
	"database/sql"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Comment struct {
	Id        uuid.UUID
	TaskId    uuid.UUID
	Value     string
	CreatedAt time.Time
	DeletedAt sql.NullTime
}
