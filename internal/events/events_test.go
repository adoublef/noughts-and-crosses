package events

import (
	"bytes"
	"encoding/gob"
	"errors"
	"testing"

	"github.com/hyphengolang/prelude/testing/is"
)

func TestDecoding(t *testing.T) {
	is := is.New(t)

	// s := "hello"
	var foo struct {
		Bar *string
	}
	// var buf bytes.Buffer
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(foo)
	is.NoErr(err) // encoding pointer

	var bar struct {
		Bar *string
	}

	err = gob.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&bar)
	is.NoErr(err) // decoding pointer

	// is.Equal(*bar.Bar, "hello") // decoding pointer
}

type foo struct {
	Value int
}

func (f *foo) MarshalBinary() ([]byte, error) {
	return []byte{}, errors.New("foo")
}

func TestEncode(t *testing.T) {
	is := is.New(t)

	f := foo{Value: 1}

	gob.Register(&foo{})

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(f)
	is.NoErr(err) // encoding pointer
}
