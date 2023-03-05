package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gobwas/ws/wsutil"
	"github.com/redis/go-redis/v9"
)

type redisPS struct {
	r         *redis.Client
	channel   string
	broadcast chan *wsutil.Message
}

func (ps *redisPS) Publish(ctx context.Context, payload *wsutil.Message) error {
	// NOTE -- this could be gob encoded
	p, err := json.Marshal(&payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return ps.r.Publish(context.Background(), ps.channel, string(p)).Err()
}

func (ps *redisPS) Subscribe(ctx context.Context) <-chan *wsutil.Message {
	return ps.broadcast
}

func (ps *redisPS) subscribe(ctx context.Context) {
	for msg := range ps.r.Subscribe(ctx, ps.channel).Channel() {
		var payload wsutil.Message
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			continue
		}
		ps.broadcast <- &payload
	}
}

func NewRedis(ctx context.Context, r *redis.Client, channel string) PubSub {
	ps := &redisPS{
		r:         r,
		channel:   channel,
		broadcast: make(chan *wsutil.Message, 256),
	}
	// go func() {
	// 	for msg := range ps.r.Subscribe(ctx, ps.channel).Channel() {
	// 		var payload wsutil.Message
	// 		// NOTE -- this could be gob encoded
	// 		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
	// 			continue
	// 		}
	// 		ps.broadcast <- &payload
	// 	}
	// }()
	go ps.subscribe(ctx)
	return ps
}
