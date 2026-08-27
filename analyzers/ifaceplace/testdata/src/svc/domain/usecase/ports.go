// Eval for GID-271 (iface-file-single-consumer), positive class: a port file
// (top-level declarations are only interfaces) whose interfaces are used by
// exactly one struct of the package — the shape of
// consent-webhook-trigger internal/domain/usecase/webhook_trigger_v2_ports.go.
// One diagnostic per file, at the first interface declaration.
package usecase

// Sender and Clock are both consumed by the single Trigger struct
// (webhook.go): one file-level violation, reported once.
type Sender interface { // want `GID-271: the file declares only interfaces, and exactly one struct \(Trigger\) in the package uses them\. Fix: move the interface declaration to the file of its consumer \(webhook\.go\)`
	Send(msg string) error
}

type Clock interface {
	Now() int64
}
