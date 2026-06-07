// Eval для GID-103 (имена ресиверов).
package service

type Snapshot struct{ name string }

type Snapshots []Snapshot

// --- Позитивные кейсы ---

func (svc *Snapshot) Bad() string { return svc.name } // want `GID-103: receiver of type Snapshot is named "s"\. Fix: use the lowercase first letter of the type \(two for slice types\), got "svc"`

func (this *Snapshot) Worse() string { return this.name } // want `GID-103: receiver of type Snapshot is named "s"`

// Граничный кейс: слайс-тип требует две буквы.
func (s Snapshots) IDs() []string { return nil } // want `GID-103: receiver of type Snapshots is named "ss"`

// --- Негативные кейсы ---

func (s *Snapshot) Name() string { return s.name }

func (ss Snapshots) Names() []string { return nil }

// Неприменимость: безымянный ресивер.
func (*Snapshot) Static() {}
