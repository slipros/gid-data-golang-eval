// Positive: no exceptions — a handler package receiver follows the same rule.
package handler

type Snapshot struct{}

func (h *Snapshot) Get() string { return "" } // want `GID-103: receiver of type Snapshot is named "s"\. Fix: use the lowercase first letter of the type \(two for slice types\), got "h"`
