package successnil

func missingUser(users []*table) (*table, error) {
	if len(users) == 0 {
		return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
	}
	return users[0], nil
}

func missingMap() (map[string]table, error) {
	return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func missingInterface() (any, error) {
	return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func missingChannel() (chan int, error) {
	return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func missingFunction() (func(), error) {
	return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func missingAlias() (tableAlias, error) {
	return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func missingGroupedResults() (first, second *table, err error) {
	return nil, nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func missingMixedResults() ([]table, *table, bool, error) {
	return nil, nil, false, nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func convertedNil() (*table, error) {
	return (*table)(nil), error(nil) // want "GID-298: non-slice nil result is returned with a nil error"
}

func convertedInterfaceNil() (any, error) {
	return any(nil), nil // want "GID-298: non-slice nil result is returned with a nil error"
}

func allowedSlice() ([]table, error) {
	return nil, nil
}

func allowedNamedSlice() (tables, error) {
	return tables(nil), nil
}

func allowedPointerError() (*table, error) {
	return nil, unavailable
}

func allowedSliceError() ([]table, error) {
	return nil, unavailable
}

func allowedPointerSuccess() (*table, error) {
	return new(table), nil
}

func allowedScalar() (int, error) {
	return 0, nil
}

func allowedSingleError() error {
	return nil
}

func allowedShadowedNil(nil *table) (*table, error) {
	return nil, unavailable
}

func typedNilInterface() (any, error) {
	return (*table)(nil), nil
}

func closureSliceOuter() ([]table, error) {
	consume(func() (*table, error) {
		return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
	})
	return nil, nil
}

func closurePointerOuter() (*table, error) {
	consume(func() ([]table, error) {
		return nil, nil
	})
	return nil, unavailable
}

func closureWithoutError() {
	consume(func() (*table, error) {
		return nil, nil // want "GID-298: non-slice nil result is returned with a nil error"
	})
}
