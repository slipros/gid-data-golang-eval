// Non-applicability (GID-176): a nested "client" segment below another layer
// (server-side connect interceptor at connect/client/interceptor) is NOT the
// client boundary layer — the layer is anchored to the first segment after the
// module root ("connect"). An interface-method call here is mechanism (b),
// which applies only inside boundary layers, so its result may be passed
// through as is without errors.Wrap.
package interceptor

type Transport interface {
	do() error
}

type Interceptor struct {
	transport Transport
}

func (i *Interceptor) passThrough() error {
	err := i.transport.do()
	return err
}
