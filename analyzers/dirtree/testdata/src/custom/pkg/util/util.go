// Positive: util is not allowed in pkg of the custom tree — the control
// works at any level, not only in internal/.
package util // want `GID-158: folder "util" is not allowed in pkg/ \(allowed: api, contract\); configure the tree via settings\.tree`

func Helper() {}
