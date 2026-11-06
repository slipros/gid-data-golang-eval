package banned

import postgres "fork.example/postgres"

func bad() {
	postgres.TQuery() // want "GID-217"
}

func good() postgres.NamedArgs {
	return nil
}
