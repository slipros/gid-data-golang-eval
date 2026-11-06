package gdhelper

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] { return make(Set[T]) }
