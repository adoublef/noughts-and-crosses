package parse

import (
	"testing"

	"github.com/hyphengolang/prelude/testing/is"
)

func TestParseDomain(t *testing.T) {
	is := is.New(t)

	url := ParseDomain("foo@gmail.com")
	is.Equal(url, "https://mail.google.com") // gmail url

	url = ParseDomain("bar@yahoo.co.uk")
	is.Equal(url, "https://mail.yahoo.com") // yahoo url

	url = ParseDomain("baz@icloud.com")
	is.Equal(url, "") // unsupported domain
}
