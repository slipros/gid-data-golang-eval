// Stub of genproto udmpapis/type/error: the MappedFields call option and its
// constructor, read by MappedFieldsUnaryClientInterceptor. The package is named
// `error` upstream too, which is why every importer aliases it —
// lk-api spells it gdgrpcerror, and so do these fixtures.
package error

import (
	gdmapper "helper/mapper"

	"google.golang.org/grpc"
)

type MappedFieldsInterceptorCallOption struct {
	grpc.EmptyCallOption

	MappedFields gdmapper.MappedFields
}

func WithMappedFields(mappedFields gdmapper.MappedFields) MappedFieldsInterceptorCallOption {
	return MappedFieldsInterceptorCallOption{MappedFields: mappedFields}
}
