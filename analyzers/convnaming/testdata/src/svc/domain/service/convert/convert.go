// Eval GID-105: converter names in a convert package.
package convert

type model struct{ Name string }

type entity struct{ Name string }

// --- Positive: a name not following the pattern ---

func ConvertSnapshot(in *model) entity { // want `GID-105: converter "ConvertSnapshot" must be named <Dst><Type>From<Src>\. Fix: rename it, e\.g\. EntityCreateSnapshotFromModel`
	return entity{Name: in.Name}
}

func ToEntity(in *model) entity { // want `GID-105: converter "ToEntity" must be named <Dst><Type>From<Src>`
	return entity{Name: in.Name}
}

// --- Negative: canonical names ---

func EntitySnapshotFromModel(in *model) entity {
	return entity{Name: in.Name}
}

func ModelSnapshotFromEntity(in *entity) model {
	return model{Name: in.Name}
}

// Not applicable: private helpers of a convert package are not checked.
func trim(s string) string { return s }
