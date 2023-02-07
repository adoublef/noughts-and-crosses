package events

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/nats-io/nats.go"
)

type Broker interface {
	Conn() *nats.Conn
	Encode(v any) ([]byte, error)
	Decode(p []byte, v any) error
	Publish(subject string, data any) error
	Subscribe(subject string, cb nats.MsgHandler)
	Request(subject string, data any, timeout time.Duration) ([]byte, error)
	RequestAndResponse(subj string, response any, request any, timeout time.Duration) error
}

var _ Broker = (*pubsub)(nil)

type pubsub struct {
	nc *nats.Conn
}

// Conn implements Broker
func (p *pubsub) Conn() *nats.Conn {
	return p.nc
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

// TODO
func (c *pubsub) RequestAndResponse(subj string, response, request any, timeout time.Duration) error {
	q, err := Encode(request)
	if err != nil {
		return err
	}

	p, err := c.nc.Request(subj, q, timeout)
	if err != nil {
		return err
	}

	return Decode(p.Data, response)
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

func Decode(p []byte, v any) error {
	return gob.NewDecoder(bytes.NewReader(p)).Decode(v)
}

func Encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
