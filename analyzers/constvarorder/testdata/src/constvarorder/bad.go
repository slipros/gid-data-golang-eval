// Eval для GID-130 — нарушения порядка.
package constvarorder

import "time"

var Late = time.Second // объявлен до const — сам по себе не нарушение...

const AfterVar = 1 // want `GID-130: const-блок размещается сверху файла — сразу после import, выше var, типов и функций`

type Svc struct{}

const AfterType = 2 // want `GID-130: const-блок размещается сверху файла — сразу после import, выше var, типов и функций`

func Do() {}

var AfterFunc = 3 // want `GID-130: var-блок размещается сверху файла — после const, выше типов и функций`
