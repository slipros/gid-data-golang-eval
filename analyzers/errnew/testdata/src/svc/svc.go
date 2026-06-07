package svc

import (
	stderrors "errors"

	"github.com/pkg/errors"

	"svc/othernew"
)

// --- Негатив: package-level var — норма ---

// ErrNotFound — статичная ошибка в одиночном var. errors.New тут легитимен.
var ErrNotFound = errors.New("not found")

// var-блок с несколькими статичными ошибками — норма.
var (
	ErrConflict = errors.New("conflict")
	ErrLocked   = errors.New("locked")
)

// --- Позитив: errors.New в func-литерале внутри package-level var ---

// makeErr — package-level var с func-литералом; errors.New в его теле
// вычисляется при вызове литерала → рантайм.
var makeErr = func() error {
	return errors.New("made at runtime") // want `GID-136: errors.New at runtime`
}

// --- Позитив: errors.New в теле функции ---

func loadSomething() error {
	return errors.New("load failed") // want `GID-136: errors.New at runtime`
}

// --- Граница: errors.Errorf в теле — не зона GID-136 ---

func formatSomething(id int) error {
	return errors.Errorf("bad id %d", id) // динамический контекст — легитимен (GID-144/145)
}

// --- Граница: стандартный errors.New в теле — зона GID-146, не GID-136 ---

func stdNew() error {
	return stderrors.New("std new") // std errors — не задеваем
}

// --- Граница: локальная функция New другого пакета — не матчится ---

func otherNew() error {
	return othernew.New("other") // не github.com/pkg/errors
}
