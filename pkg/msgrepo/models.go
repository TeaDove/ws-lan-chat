package msgrepo

import (
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Message struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`

	ConnID uuid.UUID `gorm:"not null"`
	User   string    `gorm:"not null"`
	Text   string    `gorm:"not null"`
}

func (r *Message) MarshalZerologObject(e *zerolog.Event) {
	e.Uint64("id", r.ID).
		Str("text", r.Text).
		Time("created", r.CreatedAt).
		Str("user", r.User)
}
