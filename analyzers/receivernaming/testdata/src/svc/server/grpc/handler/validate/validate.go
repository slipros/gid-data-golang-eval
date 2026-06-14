// Negative (boundary): in validate packages the receiver v is a styleguide exception.
package validate

type CreateSnapshot struct{}

func (v *CreateSnapshot) Validate() error { return nil }
