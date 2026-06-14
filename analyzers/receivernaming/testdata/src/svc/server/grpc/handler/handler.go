// Negative (boundary): in handler packages the receiver h is a styleguide exception.
package handler

type Snapshot struct{}

func (h *Snapshot) Get() string { return "" }
