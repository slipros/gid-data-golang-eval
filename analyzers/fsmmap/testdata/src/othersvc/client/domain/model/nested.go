package model

type ClientStatus string

const (
	ClientStatusNew ClientStatus = "new"
	ClientStatusOK  ClientStatus = "ok"
)

// Boundary: this package lives at othersvc/client/domain/model — the
// domain/model segments are nested under the client layer, not anchored
// right after the module root, so this is NOT the /domain/model layer.
// A substring-style Contains(path,"domain","model") check would
// false-positive on the exported transition map below; the layer-anchored
// check (HasLayer) must stay clean here.
var ClientStatusTransitions = map[ClientStatus][]ClientStatus{
	ClientStatusNew: {ClientStatusOK},
}
