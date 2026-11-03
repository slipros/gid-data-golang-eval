package errors

func Is(err, target error) bool     { return false }
func As(err error, target any) bool { return false }
