// Negative (settings.disable: GID-224): the project deliberately allowed
// transport to import domain/service — there is no diagnostic.
package handler

import "custom/domain/service"

type Snapshot struct {
	svc *service.Snapshot
}
