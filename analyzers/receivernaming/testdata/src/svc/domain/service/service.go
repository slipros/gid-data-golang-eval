// Eval for GID-103 (receiver names).
package service

type Snapshot struct{ name string }

type Snapshots []Snapshot

// --- Positive cases ---

func (svc *Snapshot) Bad() string { return svc.name } // want `GID-103: receiver of type Snapshot is named "s"\. Fix: use the lowercase first letter of the type \(two for slice types\), got "svc"`

func (this *Snapshot) Worse() string { return this.name } // want `GID-103: receiver of type Snapshot is named "s"`

// Boundary case: a slice type requires two letters.
func (s Snapshots) IDs() []string { return nil } // want `GID-103: receiver of type Snapshots is named "ss"`

// --- Negative cases ---

func (s *Snapshot) Name() string { return s.name }

func (ss Snapshots) Names() []string { return nil }

// Not applicable: an unnamed receiver.
func (*Snapshot) Static() {}
