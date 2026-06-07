// Eval для GID-175: неприменимость проверки 3 вне service/usecase.
// Анонимная tx-сигнатура в main (не слой service/usecase) — проверка 3
// не действует, диагностики нет. Проверка 1 ловит только именованные
// объявления типов; здесь анонимный тип в параметре, поэтому чисто.
package main

import "context"

func run(tx func(ctx context.Context, fn func(ctx context.Context) error) error) {
	_ = tx
}

func main() {}
