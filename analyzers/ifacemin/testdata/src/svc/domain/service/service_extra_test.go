// Use from tests does not count: SnapshotProbe.Ping still
// violates GID-197.
package service

func pingAll(p prober) error {
	return p.probe.Ping()
}

var _ = pingAll
