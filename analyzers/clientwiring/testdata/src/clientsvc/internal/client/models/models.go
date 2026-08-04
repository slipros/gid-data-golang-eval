// Eval of GID-255 boundary: a package of pure types under /client/** — the
// client's own models, no functions at all. Nothing here pretends to be a
// client, so there is nothing to move: the rule needs a function to fire.
package models

// Account is a client-side model.
type Account struct {
	ID    string
	Title string
}

// Accounts is a list of accounts.
type Accounts []Account
