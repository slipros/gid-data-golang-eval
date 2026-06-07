// Eval для GID-216: подпакет validate в consumer-scope — валидаторы,
// не consumer'ы; правило не применяется (граничный кейс).
package validate

type OrderValidator struct{}

// Конструктор валидатора без logger — НЕ должен флагаться (validate исключён).
func NewOrderValidator() *OrderValidator {
	return &OrderValidator{}
}
