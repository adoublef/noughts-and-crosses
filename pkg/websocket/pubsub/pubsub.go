package pubsub

import (
	"context"

	"github.com/gobwas/ws/wsutil"
)

type PubSub interface {
	Publish(ctx context.Context, msg *wsutil.Message) error
	Subscribe(ctx context.Context) <-chan *wsutil.Message
}

type pubsub struct {
	broadcast chan *wsutil.Message
}

func (ps *pubsub) Publish(ctx context.Context, msg *wsutil.Message) error {
	ps.broadcast <- msg
	return nil
}

func (ps *pubsub) Subscribe(ctx context.Context) <-chan *wsutil.Message {
	return ps.broadcast
}

func New(cap int) PubSub {
	ps := &pubsub{
		broadcast: make(chan *wsutil.Message, cap),
	}
	return ps
}
