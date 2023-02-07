package events

import (
	"time"

	"github.com/nats-io/nats.go"
)

type Broker interface {
	// Conn() *nats.Conn
	Subscribe(subject string, cb nats.MsgHandler)
	Publish(msg *nats.Msg) error
	Request(msg *nats.Msg, timeout time.Duration) ([]byte, error)
}

var _ Broker = (*pubsub)(nil)

type pubsub struct {
	nc *nats.Conn
}

func NewClient(nc *nats.Conn) Broker {
	return &pubsub{nc: nc}
}

func (ps *pubsub) Subscribe(subject string, cb nats.MsgHandler) {
	_, err := ps.nc.Subscribe(subject, cb)
	if err != nil {
		panic(err)
	}
}

func (ps *pubsub) Publish(msg *nats.Msg) error {
	return ps.nc.PublishMsg(msg)
}

func (ps *pubsub) Request(msg *nats.Msg, timeout time.Duration) ([]byte, error) {
	m, err := ps.nc.RequestMsg(msg, timeout)
	if err != nil {
		return nil, err
	}

	return m.Data, nil
}
