// Boundary (GID-170): the segment "event-api" contains the substring "event",
// but as a path segment is not equal to "event" — it does not fall under the rule.
package contract

type SnapshotContract struct {
	ID string
}
