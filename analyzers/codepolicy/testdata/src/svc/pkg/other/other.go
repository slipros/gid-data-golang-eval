package other

import "encoding/json"

type Alias = int

type NamedArgs map[string]any

func use(args NamedArgs) {}

func unrelated() ([]byte, error) {
	use(NamedArgs{"key": "value"})
	return json.Marshal(struct{ ID int }{ID: 1})
}

func status(value string) bool { return value == "cabinet_not_verified" } // want `GID-293:`

func arbitraryMap() map[string]struct{} {
	result := make(map[string]struct{})
	result["id"] = struct{}{}
	return result
}
