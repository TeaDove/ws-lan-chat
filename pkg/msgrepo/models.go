package msgrepo

import (
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Message struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"createdAt"`

	ConnID uuid.UUID `gorm:"not null" json:"connId"`
	User   string    `gorm:"not null" json:"user"`
	Text   string    `gorm:"not null" json:"text"`
}

func (r *Message) MarshalZerologObject(e *zerolog.Event) {
	e.Uint64("id", r.ID).
		Str("text", r.Text).
		Time("created", r.CreatedAt).
		Str("user", r.User)
}
