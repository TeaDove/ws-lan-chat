package msgrepo

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) SaveMessage(ctx context.Context, msg *Message) error {
	err := r.db.WithContext(ctx).Save(msg).Error
	if err != nil {
		return errors.Wrap(err, "save message")
	}

	zerolog.Ctx(ctx).Info().Object("msg", msg).Msg("msg.saved")

	return nil
}

func (r *Repo) ListMessages(ctx context.Context, age time.Duration, limit int) ([]Message, error) {
	msgs, err := gorm.G[Message](r.db).
		Where("created_at > ?", time.Now().Add(-age)).
		Limit(limit).
		Order("created_at desc").
		Find(ctx)

	if err != nil {
		return nil, errors.Wrap(err, "find message")
	}

	return msgs, nil
}
