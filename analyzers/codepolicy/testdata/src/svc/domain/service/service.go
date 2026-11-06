package service

import (
	"encoding/json"
	"errors"
	gderr "example.com/gderr"

	"svc/domain/model/enum"
	"svc/domain/service/convert"
)

type Model struct{ ID int }

func (m Model) Validated() Model { return m }

func inlineValidated() Model {
	return (Model{ID: 1}).Validated() // want `GID-287:`
}

func explicitValidated() Model {
	model := Model{ID: 1}
	return model.Validated()
}

func guarded() (convert.RowValue, error) {
	row, err := convert.Row(1)
	if err != nil {
		return convert.RowValue{}, err
	}
	return row, err // want `GID-288:`
}

func correct() (convert.RowValue, error) {
	row, err := convert.Row(1)
	if err != nil {
		return convert.RowValue{}, err
	}
	return row, nil
}

func failure() (convert.RowValue, error) {
	row, convErr := convert.Row(1)
	if convErr != nil {
		row.ID = 1                     // want `GID-289:`
		row.PreparationError = convErr // want `GID-289:` `GID-294:`
	}
	return row, nil
}

func correctFailure() (convert.RowValue, error) {
	row, err := convert.Row(1)
	if err != nil {
		return convert.RowValue{}, err
	}
	row.ID = 1
	return row, nil
}

func batch(values []int) []string {
	var out []string
	for _, v := range values { // want `GID-290:`
		out = append(out, convert.Value(v))
	}
	return out
}

func batchAssigned(values []int) []string {
	var out []string
	for _, v := range values { // want `GID-290:`
		converted := convert.Value(v)
		out = append(out, converted)
	}
	return out
}

func sideEffects(values []int) {
	for _, v := range values {
		_ = convert.Value(v)
	}
}

func enumDispatch(source enum.Source) (int, error) {
	if source == enum.A { // want `GID-291:`
		return 1, nil
	} else {
		return 2, nil
	}
}

func enumFilter(source enum.Source) bool {
	if source == enum.A {
		return true
	}
	return false
}

func nonEnum(source int) int {
	if source == 1 {
		return 1
	} else {
		return 2
	}
}

func anonymous() ([]byte, error) {
	return json.Marshal(struct { // want `GID-292:`
		ID int `json:"id"`
	}{ID: 1})
}

type JSONModel struct {
	ID int `json:"id"`
}

func named() ([]byte, error) {
	return json.Marshal(JSONModel{ID: 1})
}

func status(value string) bool {
	return value == "cabinet_not_verified" // want `GID-293:`
}

const cabinetNotVerified = "cabinet_not_verified"

func namedStatus(value string) bool {
	return value == cabinetNotVerified || value == "" || value == "x" // want `GID-293:`
}

var ErrFirst = errors.New("first")
var ErrSecond = errors.New("second") // want `GID-296:`

var unrelated = 1

var (
	ErrThird  = errors.New("third")
	ErrFourth = errors.New("fourth")
)

func adjacentOne() {}
func adjacentTwo() {} // want `GID-295:`

func badSwitch(source enum.Source) error {
	switch source { // want `GID-291:`
	case enum.A:
		return nil
	default:
		return nil
	}
}

func exhaustiveSwitch(source enum.Source) error {
	switch source {
	case enum.A, enum.B:
		return nil
	default:
		return gderr.NewUnhandledValueError(source)
	}
}

func marshalVariable() ([]byte, error) {
	value := struct {
		ID int `json:"id"`
	}{ID: 1}
	return json.Marshal(value) // want `GID-292:`
}

func exhaustiveLocalError(source enum.Source) error {
	switch source {
	case enum.A, enum.B:
		return nil
	default:
		err := gderr.NewUnhandledValueError(source)
		return err
	}
}

func enumBoolean(source enum.Source) bool {
	if source == enum.A {
		return true
	} else {
		return false
	}
}

func enumString(source enum.Source) string {
	switch source {
	case enum.A:
		return "a"
	default:
		return "b"
	}
}

func (m Model) Identity() Model { return m }

func immediateMethod() Model {
	return (Model{ID: 1}).Identity() // want `GID-287:`
}

func savedButReturned() (convert.RowValue, error) {
	row, convErr := convert.Row(1)
	if convErr != nil {
		row.PreparationError = convErr // want `GID-289:`
	}
	return row, convErr
}

func fallbackReplacement() (int, error) {
	row, err := convert.Row(1)
	if err != nil {
		row = convert.RowValue{ID: 2}
		return row.ID, nil
	}
	return row.ID, nil
}

func collectWithSuppressedError(values []int) ([]convert.RowValue, error) {
	var rows []convert.RowValue
	for _, value := range values { // want `GID-290:`
		row, err := convert.Row(value)
		if err != nil {
			row.PreparationError = err // want `GID-289:` `GID-294:`
		}
		rows = append(rows, row)
	}
	return rows, nil
}
