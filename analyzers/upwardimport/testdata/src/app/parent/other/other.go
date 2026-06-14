// The sibling for the negative check: importing the sibling package
// (app/parent/other) from app/parent/badchild is not a parent import.
package other

// Sibling — a type of the sibling child package.
type Sibling struct {
	ID int
}
