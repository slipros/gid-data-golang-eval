package service

type EmptyMarker interface{}

type NumberConstraint interface {
	~int | ~int64
}
