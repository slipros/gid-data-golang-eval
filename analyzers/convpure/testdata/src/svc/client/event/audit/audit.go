// Vocabulary package nested under the client layer: its layer segments are
// ["client","event","audit"], so it is NOT the banned event layer itself —
// only a layer-anchored check (pathseg.HasLayer) tells the two apart, used by
// the convpure boundary test in ../../../notify/convert/convert.go.
package audit

type Entry struct {
	ID string
}
