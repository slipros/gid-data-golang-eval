// Stub of gitlab.gid.team/gid-data/tech/golang/libs/helper.git/mapper.
package mapper

type MappedField interface {
	MapField(string) (string, bool)
}

type MappedFields []MappedField

func (mf MappedFields) MapField(value string) (string, bool) {
	for _, m := range mf {
		if v, ok := m.MapField(value); ok {
			return v, true
		}
	}

	return "", false
}

type mappedFieldString struct {
	from string
	to   string
}

func (f mappedFieldString) MapField(value string) (string, bool) {
	if value == f.from {
		return f.to, true
	}

	return "", false
}

func NewMappedFieldStringEqualWithWholePart(from, to string) MappedField {
	return mappedFieldString{from: from, to: to}
}
