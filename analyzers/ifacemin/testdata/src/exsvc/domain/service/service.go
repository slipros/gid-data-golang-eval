// Eval for GID-197: settings.exclude — "Interface" and "Interface.Method".
package service

type LegacyGateway interface {
	Send(msg string) // the interface is excluded as a whole
}

type AlertSink interface {
	Alert(msg string) // want `GID-197: method "Alert" of interface "AlertSink" is not used in the consumer package\. Fix: keep the interface minimal, remove the method`
	Flush() error     // excluded as AlertSink.Flush
}
