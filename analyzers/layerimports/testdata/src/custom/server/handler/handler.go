// Негатив (settings.disable: GID-224): проект осознанно разрешил
// транспорту импортировать domain/service — диагностики нет.
package handler

import "custom/domain/service"

type Snapshot struct {
	svc *service.Snapshot
}
