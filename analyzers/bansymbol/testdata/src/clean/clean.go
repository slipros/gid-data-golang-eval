// Class 4: not applicable — a package without an import of the banned library.
package clean

import "example.com/otherdb"

// No banned symbols: clean.
func loadOther() (int, error) {
	return otherdb.TQuery[int]("select 1")
}

func plain(x int) int { return x * 2 }
