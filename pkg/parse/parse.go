package parse

import "strings"

// https://email-verify.my-addr.com/list-of-most-popular-email-domains.php
//
// NOTE for now, we only support gmail, yahoo, hotmail, and outlook
func ParseDomain(email string) string {
	at := strings.LastIndex(email, "@")
	if at >= 0 {
		switch domain := email[at+1:]; {
		// Gmail, https://mail.google.com/mail/u/0/#inbox
		case strings.HasPrefix(domain, "gmail."):
			return "https://mail.google.com"
		case strings.HasPrefix(domain, "yahoo"):
			return "https://mail.yahoo.com"
		case strings.HasPrefix(domain, "hotmail"), strings.HasPrefix(domain, "live"), strings.HasPrefix(domain, "outlook"):
			return "https://outlook.live.com"
		default:
			return ""
		}
	} else {
		return ""
	}
}
