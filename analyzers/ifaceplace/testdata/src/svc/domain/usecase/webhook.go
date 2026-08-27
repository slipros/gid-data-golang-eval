// Eval for GID-271: the consumer side of the ports.go port file. The same
// package — GID-134 stays silent (the interfaces do live at the consumer
// package); GID-271 counts the consumers.
package usecase

// Trigger is the only struct of the package using the interfaces of
// ports.go — exactly one consumer, so the port file is a violation.
type Trigger struct {
	sender Sender
	clock  Clock
}

func (t *Trigger) Fire(msg string) error {
	return t.sender.Send(msg)
}
