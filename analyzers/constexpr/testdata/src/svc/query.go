// Eval of GID-254: a local variable initialized with a CONSTANT string
// expression that is not a lone literal (a concatenation of constants, a
// reference to a constant) must be declared as a const — the value is folded
// at compile time, nothing about it is dynamic.
package svc

import "fmt"

const (
	columns   = "id, title, created_at"
	table     = "integration"
	argID     = "id"
	maxRetry  = 3
	sqlPrefix = "SELECT "
)

// Name is a named string type — a typed constant of it is still a constant.
type Name string

const namePrefix Name = "svc_"

// --- Class 1: positive (the violation is caught) ---

// The rule's origin: a repository query assembled from constants.
func selectByID() string {
	sqlQuery := sqlPrefix + columns + " FROM " + table + " WHERE id = @id" // want `GID-254: "sqlQuery" is initialized with a constant string expression`
	return sqlQuery
}

// The same across several lines — line breaks do not make it dynamic.
func selectByOrganization() string {
	sqlQuery := "SELECT " + columns + // want `GID-254: "sqlQuery" is initialized with a constant string expression`
		" FROM " + table +
		" WHERE id = @id AND organization_id = @organization_id"
	return sqlQuery
}

// var x = <const expr> is the same declaration in a different spelling.
func selectCount() string {
	var sqlQuery = "SELECT count(*) FROM " + table // want `GID-254: "sqlQuery" is initialized with a constant string expression`
	return sqlQuery
}

// A bare reference to another constant — no concatenation needed.
func alias() string {
	cols := columns // want `GID-254: "cols" is initialized with a constant string expression`
	return cols
}

// A named string type: a typed constant expression is still constant.
func named() Name {
	full := namePrefix + "integration" // want `GID-254: "full" is initialized with a constant string expression`
	return full
}

// Several names declared at once — each is judged on its own initializer.
func pair() (string, string) {
	head, tail := sqlPrefix+columns, " FROM "+table // want `GID-254: "head" is initialized with a constant string expression` `GID-254: "tail" is initialized with a constant string expression`
	return head, tail
}

// --- Class 2: negative (clean code passes) ---

// Already a const — the shape the rule asks for.
func good() string {
	const sqlQuery = "SELECT " + columns + " FROM " + table
	return sqlQuery
}

// Genuinely dynamic: an operand is a variable, so there is no constant value.
func dynamic(tableName string) string {
	sqlQuery := "SELECT " + columns + " FROM " + tableName
	return sqlQuery
}

// Built at run time through a call — not a constant expression.
func formatted(id int) string {
	sqlQuery := fmt.Sprintf("SELECT %s FROM %s WHERE id = %d", columns, table, id)
	return sqlQuery
}

// Reassigned later — a const cannot be reassigned.
func reassigned(withFilter bool) string {
	sqlQuery := "SELECT " + columns + " FROM " + table
	if withFilter {
		sqlQuery = sqlQuery + " WHERE id = @id"
	}
	return sqlQuery
}

// Its address is taken — a const has no address.
func addressed() *string {
	sqlQuery := "SELECT " + columns
	return &sqlQuery
}

// --- Class 3: boundary (looks like a violation but is allowed) ---

// A lone string literal is deliberately out of scope: the rule targets
// expressions that LOOK assembled while being constant, not every local string.
func loneLiteral() string {
	msg := "integration not found"
	return msg
}

// The declaration is the init statement of an if — Go has no const there.
func inIfInit() bool {
	if sqlQuery := "SELECT " + columns; sqlQuery != "" {
		return true
	}
	return false
}

// A numeric constant expression is out of scope for now (string kind only).
func numeric() int {
	timeout := maxRetry * 60
	return timeout
}

// --- Class 4: non-applicability ---

// A package-level var is not a local declaration — GID-130/GID-194 own that place.
var packageQuery = "SELECT " + columns + " FROM " + table

// A short variable declaration whose value comes from a function has no
// constant value at all.
func fromCall() string {
	sqlQuery := build()
	return sqlQuery
}

func build() string { return columns }
