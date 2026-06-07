package timenow

import "time"

// stdTime — заглушка обёртки gdhelper.StdTime для негативного кейса.
var stdTime = struct{ Now func() time.Time }{Now: func() time.Time { return time.Time{} }}

// Позитив: прямой вызов time.Now() запрещён.
func bad() time.Time {
	return time.Now() // want `GID-001: time\.Now\(\) must not be called directly\. Fix: use gdhelper\.StdTime\.Now\(\) instead of time\.Now\(\)\.`
}

// Негатив: обёртка вместо time.Now().
func good() time.Time {
	return stdTime.Now()
}

// Неприменимость: одноимённый метод другого типа — не time.Now.
type clock struct{}

func (c clock) Now() time.Time { return time.Time{} }

func boundary() time.Time {
	var c clock
	return c.Now()
}
