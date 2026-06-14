// Eval for GID-216: the validate subpackage in consumer scope holds validators,
// not consumers; the rule does not apply (boundary case).
package validate

type OrderValidator struct{}

// A validator constructor without a logger must NOT be flagged (validate is excluded).
func NewOrderValidator() *OrderValidator {
	return &OrderValidator{}
}
