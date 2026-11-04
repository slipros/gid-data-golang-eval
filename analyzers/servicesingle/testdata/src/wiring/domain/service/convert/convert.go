package convert

type Converter struct{ Value int }

func (*Converter) Work()                       {}
func NewConverter(value *Converter) *Converter { return value }
