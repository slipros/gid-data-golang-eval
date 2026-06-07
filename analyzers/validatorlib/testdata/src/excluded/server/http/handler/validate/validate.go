// Не исключён — репортится даже при exclude для kafka.
package validate // want `GID-164: validate package "excluded/server/http/handler/validate" must use github\.com/raoptimus/validator\.go/v2`

func Request(raw []byte) bool { return len(raw) > 0 }
