// Eval GID-105: имена конвертеров в convert-пакете.
package convert

type model struct{ Name string }

type entity struct{ Name string }

// --- Позитив: имя не по паттерну ---

func ConvertSnapshot(in *model) entity { // want `GID-105: конвертер "ConvertSnapshot" именуется <Dst><Type>From<Src> \(например, EntityCreateSnapshotFromModel\)`
	return entity{Name: in.Name}
}

func ToEntity(in *model) entity { // want `GID-105: конвертер "ToEntity" именуется <Dst><Type>From<Src>`
	return entity{Name: in.Name}
}

// --- Негатив: канонические имена ---

func EntitySnapshotFromModel(in *model) entity {
	return entity{Name: in.Name}
}

func ModelSnapshotFromEntity(in *entity) model {
	return model{Name: in.Name}
}

// Неприменимость: приватные хелперы convert-пакета не проверяются.
func trim(s string) string { return s }
