// Негатив (граница): в handler-пакетах ресивер h — исключение стайлгайда.
package handler

type Snapshot struct{}

func (h *Snapshot) Get() string { return "" }
