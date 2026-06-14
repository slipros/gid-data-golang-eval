// Not applicable: a convert package without enum map indexings — no diagnostics.
package convert

type model struct{ Name string }

type entity struct{ Name string }

// An ordinary field converter — no maps at all.
func EntityNameFromModel(in *model) entity {
	return entity{Name: in.Name}
}

func ModelNameFromEntity(in *entity) model {
	return model{Name: in.Name}
}
