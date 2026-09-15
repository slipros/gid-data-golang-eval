package service

func (h *Hello) Ping() {}

type PingNotifier interface { // want `GID-276: interface PingNotifier is declared below func \(\*Hello\) Ping \(line 3\)`
	Notify()
}
