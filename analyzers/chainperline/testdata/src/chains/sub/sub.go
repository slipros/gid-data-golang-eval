// Вспомогательный пакет для eval GID-196: конверсия через селектор.
package sub

type Code string

func (c Code) Upper() string { return string(c) }
