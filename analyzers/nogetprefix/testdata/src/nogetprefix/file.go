// Eval для GID-101 (no-get-prefix).
package nogetprefix

type Job struct {
	id     string
	status string
}

// --- Позитивные кейсы: нарушение ловится ---

func (j *Job) GetID() string { return j.id } // want `GID-101: метод "GetID" использует префикс Get — геттеры именуются без него: "ID"`

func (j *Job) GetStatus() string { return j.status } // want `GID-101: метод "GetStatus" использует префикс Get`

// Граничный кейс: голое Get — тоже нарушение.
func (j *Job) Get() string { return j.id } // want `GID-101: метод "Get" использует префикс Get`

// --- Негативные кейсы: чистый код проходит ---

func (j *Job) ID() string { return j.id }

func (j *Job) Status() string { return j.status }

// Граничный кейс: Get — часть слова, не префикс.
func (j *Job) Getaway() string { return j.id }

// --- Неприменимость: правило действует только на методы ---

func GetEnv(key string) string { return key }
