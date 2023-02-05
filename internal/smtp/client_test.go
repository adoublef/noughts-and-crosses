// https://asitdhal.medium.com/golang-sendmail-sending-mail-through-net-smtp-package-5cadbe2670e0
package smtp

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hyphengolang/prelude/testing/is"
)

var (
	host   = "smtp.gmail.com"
	port   = 587
	sender = "kristopherab@campus.masterschool.com"
	passw  = "qgwujbkhdzhnpxdz"
)

func TestSendTLS(t *testing.T) {
	is := is.New(t)

	t.Run("send flow", func(t *testing.T) {
		m := NewMailer(sender, passw, host, port)

		to := []string{"kristopherab@gmail.com"}
		body := `
		<html>

<body>
    <h1>Thank you for registering with us</h1>
    <a href="https://www.google.com">Click here to verify your email address</a>
</body>

</html>
		`

		mail := &Mail{
			// NOTE could use mail.Address instead of string
			To: to,
			// NOTE could use mail.Address instead of string
			// CC:   cc,
			Subj: "Confirm your email for your account: " + uuid.New().String(),
			Body: []byte(body),
		}

		err := m.Send(mail)
		is.NoErr(err) // mail sent
	})
}
