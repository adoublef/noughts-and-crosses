package events

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/nats-io/nats.go"
)

type Client struct {
	e events.Broker

	nc *nats.Conn
}

func (c *Client) Foo(response, request any) error {
	q, err := Encode(request)
	if err != nil {
		return err
	}

	p, err := c.nc.Request("hello", q, 5*time.Second)
	if err != nil {
		return err
	}

	return Decode(p.Data, response)
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
