package rand

import (
	"testing"

	"github.com/hyphengolang/prelude/testing/is"
)

func TestRandomString(t *testing.T) {
	is := is.New(t)

	// Test that the random string is the correct length
	rs := RandString{Length: 6}

	is.Equal(len(rs.ToUpper()), 6)
}
