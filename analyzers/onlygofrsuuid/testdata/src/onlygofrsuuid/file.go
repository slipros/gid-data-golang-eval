// Eval для GID-137 (only-gofrs-uuid).
package onlygofrsuuid

import (
	"github.com/gofrs/uuid"

	googleuuid "github.com/google/uuid" // want `GID-137: importing "github.com/google/uuid" is forbidden\. Fix: use github.com/gofrs/uuid for UUID`

	satori "github.com/satori/go.uuid" // want `GID-137: importing "github.com/satori/go.uuid" is forbidden\. Fix: use github.com/gofrs/uuid for UUID`

	"example.com/uuidutil"
)

// --- Позитив: запрещённые импорты пойманы выше (включая граничный кейс с алиасом) ---

func bad() googleuuid.UUID { return googleuuid.New() }

func badSatori() satori.UUID { return satori.NewV4() }

// --- Негатив: разрешённая библиотека проходит ---

func good() uuid.UUID { return uuid.Must(uuid.NewV7()) }

// --- Неприменимость: пакет с "uuid" в имени, но не uuid-библиотека ---

func notApplicable() string { return uuidutil.Normalize("x") }
