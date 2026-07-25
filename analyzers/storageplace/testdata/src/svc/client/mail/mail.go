// Negative: a client of an external API (SMTP) — not a storage driver.
package mail

import "github.com/wneessen/go-mail"

type Sender struct {
	client *mail.Client
}
