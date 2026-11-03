// Positive: no exceptions — a validate package receiver follows the same rule.
package validate

type CreateSnapshot struct{}

func (v *CreateSnapshot) Validate() error { return nil } // want `GID-103: receiver of type CreateSnapshot is named "c"\. Fix: use the lowercase first letter of the type \(first two word initials for slices; double the initial for one word\), got "v"`
