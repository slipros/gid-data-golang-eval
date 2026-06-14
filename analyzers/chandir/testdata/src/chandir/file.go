// Eval for GID-189 (Google: channel direction).
package chandir

// --- Positive cases (a bidirectional channel parameter) ---

// A function with a channel parameter without a direction.
func consume(ch chan int) { // want `GID-189: channel parameter ch is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
	<-ch
}

type worker struct{}

// A method with a channel parameter without a direction.
func (w worker) run(ch chan string) { // want `GID-189: channel parameter ch is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
	ch <- "x"
}

// A function literal with a channel parameter without a direction.
func withLit() {
	f := func(ch chan int) { // want `GID-189: channel parameter ch is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
		<-ch
	}
	_ = f
}

// Several names in one parameter group — one diagnostic per group.
func multi(a, b chan int) { // want `GID-189: channel parameter a, b is bidirectional\. Fix: declare a direction, <-chan to receive or chan<- to send\.`
	<-a
	<-b
}

// --- Negative cases (the direction is declared) ---

// Receive only.
func recvOnly(ch <-chan int) {
	<-ch
}

// Send only.
func sendOnly(ch chan<- int) {
	ch <- 1
}

// --- Edge cases (not matched) ---

// Returning a bidirectional channel — the owner sometimes needs it, review decides.
func produce() chan int {
	return make(chan int)
}

// A struct field with a channel type — the direction is set when passing it.
type holder struct {
	ch chan int
}

// A local channel variable — that is channel creation, not a signature.
func local() {
	var ch chan int
	_ = ch
}

// A named channel type in parameter position — a deliberate decision.
type Pipe chan int

func namedParam(p Pipe) {
	<-p
}

// A slice of channels — not a direct channel parameter.
func sliceParam(chs []chan int) {
	_ = chs
}
