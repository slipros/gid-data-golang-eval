// Eval для GID-175: проверка 1 действует везде, кроме /domain/model.
// helper не входит в слои service/usecase/repository, но проверка 1
// всё равно флагует объявление tx-типа здесь.
package helper

import "context"

// --- Проверка 1 (позитив): именованный tx-тип объявлен вне /domain/model ---

type Tx func(ctx context.Context, fn func(ctx context.Context) error) error // want `GID-175: the transaction type must live in /domain/model \(InTransactionFunc\)\. Fix: move it there`

// --- Проверка 1 (позитив): generic-вариант вне model ---

type TxR[T any] func(ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) // want `GID-175: the transaction type must live in /domain/model \(InTransactionFunc\)\. Fix: move it there`

// Граничный кейс: похожая, но другая сигнатура (без ctx) — не флагуем.
type NotTx func(fn func(ctx context.Context) error) error
