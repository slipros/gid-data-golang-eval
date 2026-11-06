package convert

func Row(value int) (RowValue, error) { return RowValue{ID: value}, nil }

type RowValue struct {
	ID               int
	PreparationError error
}

func Value(value int) string { return "" }

func Rows(values []int) ([]RowValue, error) {
	result := make([]RowValue, 0, len(values))
	for _, value := range values {
		row, err := Row(value)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}
