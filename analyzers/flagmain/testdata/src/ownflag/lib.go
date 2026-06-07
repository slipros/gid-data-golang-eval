// Класс «граничный»: локальный пакет с именем flag — это НЕ stdlib "flag",
// путь импорта другой, поэтому правило не срабатывает даже в библиотеке.
package ownflag

import "ownflag/flag"

func use() {
	flag.String("maxRetries") // имя не snake_case, но это чужой flag — диагностики нет
	flag.Parse()
}
