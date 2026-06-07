// Использование из тестов не считается: SnapshotProbe.Ping всё равно
// нарушает GID-197.
package service

func pingAll(p prober) error {
	return p.probe.Ping()
}

var _ = pingAll
