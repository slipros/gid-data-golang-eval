// Package local — constructors judged for the boundary cases of GID-268.
package local

import (
	"io"

	"svc/remote"
)

// Fetcher — one dependency interface.
type Fetcher interface {
	Fetch(id string) string
}

// Storer — another dependency interface of the same package.
type Storer interface {
	Store(id string, body string) error
}

// Listener — the interface of the variadic tail.
type Listener interface {
	OnEvent(name string)
}

// Anything — an interface with no methods: nothing to merge.
type Anything interface{}

// Pair holds two collaborators of the same interface.
type Pair struct {
	primary Fetcher
	backup  Fetcher
}

// newPair takes the same interface twice — the duplicate is a wiring question,
// not an interface one, so the rule stays silent.
func newPair(primary, backup Fetcher) *Pair {
	return &Pair{primary: primary, backup: backup}
}

// Span — a numeric range.
type Span struct {
	from int
	to   int
}

// newSpan takes no interface at all.
func newSpan(from, to int) *Span {
	return &Span{from: from, to: to}
}

// Pipe — a copier over stdlib interfaces.
type Pipe struct {
	dst io.Writer
	src io.Reader
}

// newPipe takes stdlib interfaces: passing one value under Writer and Reader is
// the intent of the shape, and neither interface is ours to merge.
func newPipe(dst io.Writer, src io.Reader) *Pipe {
	return &Pipe{dst: dst, src: src}
}

// Mixed holds interfaces declared in different packages.
type Mixed struct {
	fetcher Fetcher
	auditor remote.Auditor
}

// newMixed cannot be fixed by merging: the interfaces live in two packages.
func newMixed(fetcher Fetcher, auditor remote.Auditor) *Mixed {
	return &Mixed{fetcher: fetcher, auditor: auditor}
}

// Blank holds two method-less interfaces.
type Blank struct {
	left  Anything
	right Anything
}

// newBlank takes interfaces carrying no method — no dependency to merge.
func newBlank(left, right Anything) *Blank {
	return &Blank{left: left, right: right}
}

// Bus holds a fetcher and its listeners.
type Bus struct {
	fetcher   Fetcher
	listeners []Listener
}

// newBus takes a single fixed parameter and a variadic tail: the tail holds no
// parameter of its own to name in a fix.
func newBus(fetcher Fetcher, listeners ...Listener) *Bus {
	return &Bus{fetcher: fetcher, listeners: listeners}
}

// Feed holds two dependencies and a listener tail.
type Feed struct {
	fetcher   Fetcher
	storer    Storer
	listeners []Listener
}

// newFeed carries a variadic tail after two named dependencies.
func newFeed(fetcher Fetcher, storer Storer, listeners ...Listener) *Feed {
	return &Feed{fetcher: fetcher, storer: storer, listeners: listeners}
}

// both hands out the two dependencies at once.
func both(store *Store) (Fetcher, Storer) { return store, store }

// Store — an entity satisfying every interface used here.
type Store struct{}

// Fetch — the fetching role.
func (s *Store) Fetch(_ string) string { return "" }

// Store — the storing role.
func (s *Store) Store(_ string, _ string) error { return nil }

// OnEvent — the listening role.
func (s *Store) OnEvent(_ string) {}

// Audit — the auditing role.
func (s *Store) Audit(_ string) {}

// Write — the writing role.
func (s *Store) Write(p []byte) (int, error) { return len(p), nil }

// Read — the reading role.
func (s *Store) Read(_ []byte) (int, error) { return 0, io.EOF }

// Vault holds two dependencies of this package.
type Vault struct {
	fetcher Fetcher
	storer  Storer
}

// newVault is the shape the rule is about, inside one package.
func newVault(fetcher Fetcher, storer Storer) *Vault {
	return &Vault{fetcher: fetcher, storer: storer}
}

// other — a second entity, for the storing role only.
type other struct{}

// Store — the storing role.
func (other) Store(_ string, _ string) error { return nil }

// Wire builds everything from a single entity.
func Wire(store *Store) {
	_ = newPair(store, store)
	_ = newSpan(1, 1)
	_ = newPipe(store, store)
	_ = newMixed(store, store)
	_ = newBlank(store, store)
	_ = newBus(store, store, store)
	_ = newFeed(store, other{}, store, store)
	_ = newVault(both(store))
	_ = newVault(store, store) // want `GID-268: constructor newVault receives the same value in parameters #1 fetcher and #2 storer — one dependency passed twice under different interfaces\. Fix: merge Fetcher and Storer into a single interface and take the dependency as one parameter`
}
