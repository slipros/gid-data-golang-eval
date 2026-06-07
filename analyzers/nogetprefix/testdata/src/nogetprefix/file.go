// Eval для GID-101 (no-get-prefix).
package nogetprefix

type Job struct {
	id     string
	status string
}

// --- Позитивные кейсы: нарушение ловится ---

func (j *Job) GetID() string { return j.id } // want `GID-101: method "GetID" uses the Get prefix\. Fix: name getters without it: "ID"`

func (j *Job) GetStatus() string { return j.status } // want `GID-101: method "GetStatus" uses the Get prefix`

// Граничный кейс: голое Get — тоже нарушение.
func (j *Job) Get() string { return j.id } // want `GID-101: method "Get" uses the Get prefix`

// --- Негативные кейсы: чистый код проходит ---

func (j *Job) ID() string { return j.id }

func (j *Job) Status() string { return j.status }

// Граничный кейс: Get — часть слова, не префикс.
func (j *Job) Getaway() string { return j.id }

// --- Неприменимость: правило действует только на методы ---

func GetEnv(key string) string { return key }
