package repository

import (
	"convertsvc/dal/repository/build"
	"convertsvc/dal/repository/convert"
	"encoding/json"
	"github.com/pkg/errors"
)

func wrapped() (int, error) {
	row, err := convert.Row()
	if err != nil {
		return 0, errors.Wrap(err, "convert") // want `GID-176: repository conversion/build error`
	}
	return row, nil
}

func wrappedf() (int, error) {
	row, err := convert.Row()
	if err != nil {
		return 0, errors.Wrapf(err, "convert %d", row) // want `GID-176: repository conversion/build error`
	}
	return row, nil
}

func stacked() (string, error) {
	query, err := build.SQL()
	if err != nil {
		return "", errors.WithStack(err) // want `GID-176: repository conversion/build error`
	}
	return query, nil
}

func message() (int, error) {
	row, err := convert.Row()
	if err != nil {
		return 0, errors.WithMessage(err, "convert")
	}
	return row, nil
}

func external() error {
	var row int
	err := json.Unmarshal([]byte("1"), &row)
	if err != nil {
		return errors.Wrap(err, "unmarshal")
	}
	return nil
}

func reassigned() (int, error) {
	row, err := convert.Row()
	err = json.Unmarshal([]byte("1"), &row)
	if err != nil {
		return 0, errors.Wrap(err, "external")
	}
	return row, nil
}

func conditionalProvenance(cond bool) error {
	var row int
	err := json.Unmarshal([]byte("1"), &row)
	if cond {
		_, err = convert.Row()
		return err
	}
	return errors.Wrap(err, "unmarshal")
}

func conditionalExternal(cond bool) error {
	_, err := convert.Row()
	if cond {
		err = json.Unmarshal([]byte("1"), new(int))
	}
	return errors.Wrap(err, "ambiguous")
}

func loopPostProvenance(condition bool) error {
	var row int
	err := json.Unmarshal([]byte("1"), &row)
	for ; condition; _, err = convert.Row() {
		return errors.Wrap(err, "unmarshal")
	}
	return nil
}

func loopRepeatedProvenance(condition bool) error {
	_, err := convert.Row()
	for condition {
		if condition {
			return errors.Wrap(err, "ambiguous loop")
		}
		err = json.Unmarshal([]byte("1"), new(int))
	}
	return nil
}

func rangeErrorProvenance(externalErrors []error) error {
	_, err := convert.Row()
	for _, err = range externalErrors {
		return errors.Wrap(err, "external")
	}
	return errors.Wrap(err, "ambiguous after range")
}

func receivedErrorProvenance(externalErrors <-chan error) error {
	_, err := convert.Row()
	select {
	case err = <-externalErrors:
		return errors.Wrap(err, "external")
	default:
		return errors.WithMessage(err, "convert")
	}
}

func rangeAssignmentProvenance(externalErrors []error) error {
	_, err := convert.Row()
	for _, err = range externalErrors {
		return errors.Wrap(err, "external")
	}
	return errors.Wrap(err, "possibly external")
}

func escapedError() error {
	_, err := convert.Row()
	replaceError(&err)
	return errors.Wrap(err, "external")
}

func escapedBeforeConversion() error {
	var err error
	alias := &err
	_, err = convert.Row()
	replaceError(alias)
	return errors.Wrap(err, "external")
}

func closureReassignment() error {
	_, err := convert.Row()
	func() {
		err = json.Unmarshal([]byte("bad"), new(int))
	}()
	return errors.Wrap(err, "external")
}

func replaceError(dst *error) {
	*dst = json.Unmarshal([]byte("bad"), new(int))
}
