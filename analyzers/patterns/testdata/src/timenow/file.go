package timenow

import "time"

// stdTime — a stub of the gdhelper.StdTime wrapper for the negative case.
var stdTime = struct{ Now func() time.Time }{Now: func() time.Time { return time.Time{} }}

// Positive: a direct time.Now() call is forbidden.
func bad() time.Time {
	return time.Now() // want `GID-001: time\.Now\(\) must not be called directly\. Fix: use gdhelper\.StdTime\.Now\(\) instead of time\.Now\(\)\.`
}

// Negative: the wrapper instead of time.Now().
func good() time.Time {
	return stdTime.Now()
}

// Not applicable: a same-named method of another type — not time.Now.
type clock struct{}

func (c clock) Now() time.Time { return time.Time{} }

func boundary() time.Time {
	var c clock
	return c.Now()
}
