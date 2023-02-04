package events

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/nats-io/nats.go"
)

type Event int

const (
	EventUnknown Event = iota
	EventUserLogin
	EventTokenGenerate
	EventTokenVerify
)

func (e Event) String() string {
	switch e {
	case EventUserLogin:
		return "user.login"
	case EventTokenGenerate:
		return "token.generate"
	case EventTokenVerify:
		return "token.verify"
	default:
		return "unknown"
	}
}

type Broker interface {
	Encode(v any) ([]byte, error)
	Decode(p []byte, v any) error
	Publish(subject string, data any) error
	Subscribe(subject string, cb nats.MsgHandler)
	Request(subject string, data any, timeout time.Duration) ([]byte, error)
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

func (ps *pubsub) Publish(subject string, data any) error {
	p, err := ps.Encode(data)
	if err != nil {
		return err
	}

	return ps.nc.Publish(subject, p)
}

func (ps *pubsub) Request(subject string, data any, timeout time.Duration) ([]byte, error) {
	p, err := ps.Encode(data)
	if err != nil {
		return nil, err
	}

	msg, err := ps.nc.Request(subject, p, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return msg.Data, nil
}

func (*pubsub) Decode(p []byte, v any) error {
	return gob.NewDecoder(bytes.NewReader(p)).Decode(v)
}

func (s *pubsub) Encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
