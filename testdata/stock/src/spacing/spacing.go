package spacing

func bad(convert func() (int, error), repo func(int) (bool, error)) (bool, error) {
	row, err := convert()
	if err != nil {
		return false, err
	} // want "missing whitespace below this line"
	created, err := repo(row)
	if err != nil {
		return false, err
	} // want "missing whitespace below this line"
	return created, nil
}

func badDeclaration(exec func(string, map[string]any) error) error {
	const query = `INSERT INTO test VALUES (@id)` // want "missing whitespace below this line"
	args := map[string]any{"id": 1}

	if err := exec(query, args); err != nil {
		return err
	} // want "missing whitespace below this line"
	return nil
}

func good(convert func() (int, error), repo func(int) (bool, error)) (bool, error) {
	row, err := convert()
	if err != nil {
		return false, err
	}

	created, err := repo(row)
	if err != nil {
		return false, err
	}

	return created, nil
}

func simple(value int) int {
	return value
}
