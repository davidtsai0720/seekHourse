// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package sqlc

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type City struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	DeletedAt sql.NullTime
}

type Hourse struct {
	ID          int64
	UniversalID uuid.UUID
	SectionID   int32
	Link        string
	Layout      sql.NullString
	Address     sql.NullString
	Price       string
	Floor       string
	Shape       string
	Age         string
	Area        string
	MainArea    sql.NullString
	Raw         json.RawMessage
	Others      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

type Section struct {
	ID        int64
	CityID    int32
	Name      string
	CreatedAt time.Time
	DeletedAt sql.NullTime
}
