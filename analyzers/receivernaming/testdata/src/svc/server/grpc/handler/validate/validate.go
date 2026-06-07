// Негатив (граница): в validate-пакетах ресивер v — исключение стайлгайда.
package validate

type CreateSnapshot struct{}

func (v *CreateSnapshot) Validate() error { return nil }
