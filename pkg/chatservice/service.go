package chatservice

import (
	"context"
	"sync"
	"time"
	"ws-lan-chat/pkg/msgrepo"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Service struct {
	msgRepo *msgrepo.Repo

	connsMu sync.RWMutex
	conns   map[uuid.UUID]chan msgrepo.Message
}

func NewService(msgRepo *msgrepo.Repo) *Service {
	return &Service{msgRepo: msgRepo, conns: make(map[uuid.UUID]chan msgrepo.Message)}
}

type MsgRequest struct {
	Text string `json:"text"`
}

func (r *Service) SaveMessage(ctx context.Context, user string, connID uuid.UUID, msgRequest *MsgRequest) error {
	msg := msgrepo.Message{
		User:   user,
		ConnID: connID,
		Text:   msgRequest.Text,
	}

	err := r.msgRepo.SaveMessage(ctx, &msg)
	if err != nil {
		return errors.Wrap(err, "save message")
	}

	r.connsMu.RLock()
	defer r.connsMu.RUnlock()
	for _, ch := range r.conns {
		ch <- msg
	}

	return nil
}

func (r *Service) Subscribe(ctx context.Context, connID uuid.UUID) (chan msgrepo.Message, error) {
	r.connsMu.Lock()
	defer r.connsMu.Unlock()

	channel := make(chan msgrepo.Message, 100)
	r.conns[connID] = channel

	msgs, err := r.msgRepo.ListMessages(ctx, 7*24*time.Hour, 100)
	if err != nil {
		return nil, errors.Wrap(err, "list messages")
	}

	for _, msg := range msgs {
		channel <- msg
	}

	return channel, nil
}

func (r *Service) Unsubscribe(connID uuid.UUID) {
	r.connsMu.Lock()
	defer r.connsMu.Unlock()

	channel, ok := r.conns[connID]
	if !ok {
		return
	}

	close(channel)
	delete(r.conns, connID)
}
