// Package model here lives at svc/client/domain/model — nested under the
// client layer, not the domain/model layer itself. Used by the modelmethod
// boundary test in svc/domain/service/boundary_nested.go.
package model

type Profile struct {
	Name string
}
