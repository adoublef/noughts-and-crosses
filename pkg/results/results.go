package results

import (
	"encoding/gob"
	"fmt"
	"log"

	"github.com/hyphengolang/noughts-and-crosses/internal/events"
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

type Result[T any] struct {
	Value T
	Err   error
}

func New[T any](v T) Result[T] {
	return Result[T]{Value: v}
}

// Returns a byte slice of the result, with the error encoded
// Converted using gob encoding. The `%w` directive not allowed
func (r *Result[T]) Errorf(format string, a ...any) []byte {
	r.Err = &Err{fmt.Sprintf(format, a...)}
	return r.Bytes()
}

// FIXME I am being naughty and not error handling
// If this causes an error then
func (r Result[T]) Msg(subj string) *nats.Msg {
	p, _ := events.Encode(r)
	return &nats.Msg{Subject: subj, Data: p}
}

// Encodes the result into a byte slice
// using gob encoding
func (r Result[T]) Bytes() []byte {
	p, err := events.Encode(r)
	if err != nil {
		log.Print(err)
	}
	return p
}
