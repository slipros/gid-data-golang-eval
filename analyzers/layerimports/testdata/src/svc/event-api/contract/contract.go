// Граничный (GID-170): сегмент "event-api" содержит подстроку "event",
// но как сегмент пути не равен "event" — под правило не подпадает.
package contract

type SnapshotContract struct {
	ID string
}
