package events

import (
	"github.com/nats-io/nats.go"
)

type Broker interface {
	Conn() *nats.EncodedConn
}

var _ Broker = (*pubsub)(nil)

type pubsub struct {
	ec *nats.EncodedConn
}

func NewClient(ec *nats.EncodedConn) Broker {
	return &pubsub{ec: ec}
}

func (ps *pubsub) Conn() *nats.EncodedConn { return ps.ec }

// func (ps *pubsub) Subscribe(subject string, cb nats.MsgHandler) {
// 	_, err := ps.ec.Subscribe(subject, cb)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func (ps *pubsub) Publish(msg *nats.Msg) error {
// 	return ps.ec.Conn.PublishMsg(msg)
// }

// func (ps *pubsub) Request(msg *nats.Msg, timeout time.Duration) ([]byte, error) {
// 	m, err := ps.ec.Conn.RequestMsg(msg, timeout)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return m.Data, nil
// }
