// Package promo — a subpackage of /domain/model: the rule applies here too.
package promo

import "errors"

// Positive: a generic name in a model subpackage.
var ErrNoResult = errors.New("no result") // want `GID-234: generic error name "ErrNoResult" in domain model\. Fix: bind it to the entity: ErrSnapshotNoResult`

// Negative: an entity-bound name.
var ErrPromoNotFound = errors.New("promo not found")
