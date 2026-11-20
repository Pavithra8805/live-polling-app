package realtime

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	store *RedisStore
	hub   *Hub
}

func NewSubscriber(store *RedisStore, hub *Hub) *Subscriber {
	return &Subscriber{store: store, hub: hub}
}

func (s *Subscriber) Run(ctx context.Context) error {
	pubsub := s.store.Subscribe(ctx)
	defer pubsub.Close()
	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}
	channel := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, ok := <-channel:
			if !ok {
				return nil
			}
			pollID := pollIDFromChannel(message.Channel)
			if pollID != "" {
				s.hub.Broadcast(pollID, []byte(message.Payload))
			}
		}
	}
}

func pollIDFromChannel(channel string) string {
	const prefix = "poll:"
	const suffix = ":updates"
	if !strings.HasPrefix(channel, prefix) || !strings.HasSuffix(channel, suffix) {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(channel, prefix), suffix)
}
