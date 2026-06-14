// Not excluded — reported even with the kafka exclude in place.
package validate // want `GID-164: validate package "excluded/server/http/handler/validate" must use github\.com/raoptimus/validator\.go/v2`

func Request(raw []byte) bool { return len(raw) > 0 }
