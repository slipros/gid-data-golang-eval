// Минимальный stub github.com/gofrs/uuid для eval: ruleguard проверяет
// тип через Where(...Type.Is(`uuid.UUID`)), реальная библиотека не нужна.
package uuid

type UUID [16]byte

func (u UUID) IsNil() bool {
	return u == UUID{}
}

func (u UUID) String() string {
	return ""
}

func NewV1() (UUID, error) {
	return UUID{}, nil
}

func NewV3(ns UUID, name string) UUID {
	return UUID{}
}

func NewV4() (UUID, error) {
	return UUID{}, nil
}

func NewV5(ns UUID, name string) UUID {
	return UUID{}
}

func NewV6() (UUID, error) {
	return UUID{}, nil
}

func NewV7() (UUID, error) {
	return UUID{}, nil
}

func FromString(s string) (UUID, error) {
	return UUID{}, nil
}

func Must(u UUID, err error) UUID {
	if err != nil {
		panic(err)
	}
	return u
}
