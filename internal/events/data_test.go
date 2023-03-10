package events_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/hyphengolang/noughts-and-crosses/internal/events"
	"github.com/hyphengolang/prelude/testing/is"
)

func TestEncoding(t *testing.T) {
	is := is.New(t)

	t.Run("result with empty struct", func(t *testing.T) {
		type Data struct{ events.Data[struct{}] }
		var data Data

		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(data)
		is.NoErr(err) // encoding result type

		p := buf.Bytes()
		err = gob.NewDecoder(bytes.NewReader(p)).Decode(&data)
		is.NoErr(err) // decoding result type

		is.True(data.Value == struct{}{}) // value is struct{}{}
	})

	t.Run("result.Bytes happy", func(t *testing.T) {
		type Data struct{ events.Data[int] }

		var input Data
		input.Value = 10 // value is 0
		// been gob encoded
		p := input.Bytes()

		var output Data
		err := gob.NewDecoder(bytes.NewReader(p)).Decode(&output)
		is.NoErr(err) // decoding result type

		is.Equal(output.Value, 10) // value is struct{}{}
		is.Equal(output.Err, nil)  // error is nil
	})

	t.Run("result.Bytes error", func(t *testing.T) {
		type Data struct{ events.Data[struct{}] }

		var input Data
		p := input.Errorf("test error")

		is.True(len(p) > 0)       // bytes are returned
		is.True(input.Err != nil) // error is set

		var output Data
		err := gob.NewDecoder(bytes.NewReader(p)).Decode(&output)
		is.NoErr(err) // decoding result type

		is.True(output.Value == struct{}{}) // value is struct{}{}
		is.True(output.Err != nil)          // error is set
	})

	t.Run("result.Bytes error", func(t *testing.T) {
		type Data struct{ events.Data[struct{}] }

		var input Data
		p := input.Errorf("test error")

		is.True(len(p) > 0)       // bytes are returned
		is.True(input.Err != nil) // error is set

		var output Data
		err := gob.NewDecoder(bytes.NewReader(p)).Decode(&output)
		is.NoErr(err) // decoding result type
		is.Equal(output.Err.Error(), "test error")
	})
}
