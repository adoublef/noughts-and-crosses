package events

import (
	"encoding/gob"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func init() {
	gob.Register(Err{})
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Err struct {
	Msg string
}

func (e Err) Error() string {
	return e.Msg
}

type Data[T any] struct {
	Value T
	Err   error
}

// BErrof is a byte slice encoding of the result using gob encoding.
// The `%w` directive not allowed, as uses `fmt.Sprintf` under the hood
func (d *Data[T]) Errorf(format string, a ...any) []byte {
	d.Err = &Err{fmt.Sprintf(format, a...)}
	return d.Bytes()
}

// FIXME I am being naughty and not error handling
// If this causes an error then
func (r Data[T]) Msg(subj string) *nats.Msg {
	p, err := Marshal(r)
	if err != nil {
		log.Print(err)
	}
	return &nats.Msg{Subject: subj, Data: p}
}

// Encodes the result into a byte slice
// using gob encoding
func (d Data[T]) Bytes() []byte {
	p, err := Marshal(d)
	if err != nil {
		log.Print(err)
	}
	return p
}
