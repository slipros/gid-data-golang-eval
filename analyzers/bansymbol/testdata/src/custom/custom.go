package custom

import "example.com/otherdb"

// settings.symbols задаёт кастомный символ otherdb.TQuery с кастомным Msg —
// флагается именно этим сообщением (а дефолтный gdpostgres.TQuery в этом
// прогоне не задан).

func callBanned() (int, error) {
	return otherdb.TQuery[int]("select 1") // want `GID-217: otherdb\.TQuery is banned by the project`
}
