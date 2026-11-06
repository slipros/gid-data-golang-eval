package repository

import postgres "fork.example/postgres"

func exec(_ postgres.NamedArgs) {}

func bad() {
	exec(postgres.NamedArgs{"id": 1}) // want "GID-285"
}

func good() {
	args := postgres.NamedArgs{"id": 1}
	exec(args)
}
