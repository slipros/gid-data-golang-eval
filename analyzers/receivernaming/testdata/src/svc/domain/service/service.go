// Eval для GID-103 (имена ресиверов).
package service

type Snapshot struct{ name string }

type Snapshots []Snapshot

// --- Позитивные кейсы ---

func (svc *Snapshot) Bad() string { return svc.name } // want `GID-103: ресивер типа Snapshot именуется "s" — первая буква типа в нижнем регистре \(две для слайс-типов\), получено "svc"`

func (this *Snapshot) Worse() string { return this.name } // want `GID-103: ресивер типа Snapshot именуется "s"`

// Граничный кейс: слайс-тип требует две буквы.
func (s Snapshots) IDs() []string { return nil } // want `GID-103: ресивер типа Snapshots именуется "ss"`

// --- Негативные кейсы ---

func (s *Snapshot) Name() string { return s.name }

func (ss Snapshots) Names() []string { return nil }

// Неприменимость: безымянный ресивер.
func (*Snapshot) Static() {}
