// Eval for GID-216: outside the event layer the rule does not apply (inapplicability).
package service

type Service struct{}

// A constructor without a logger outside the event layer is NOT flagged.
func NewService() *Service {
	return &Service{}
}
