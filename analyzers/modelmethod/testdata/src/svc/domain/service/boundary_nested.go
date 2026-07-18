// Boundary (GID-195): the parameter type below lives in
// svc/client/domain/model — the domain/model segments are nested under the
// client layer, not anchored right after the module root, so it is NOT a
// domain/model-layer type. A substring-style Contains(path,"domain","model")
// check would false-positive and flag this private function as movable
// model behaviour; the layer-anchored check must leave it alone.
package service

import clientmodel "svc/client/domain/model"

func describeProfile(p *clientmodel.Profile) string {
	return p.Name
}
