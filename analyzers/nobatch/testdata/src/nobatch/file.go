// Eval для GID-102 (no-batch-word).
package nobatch

type Job struct{ id string }

type CreateJob struct{ ID string }

// --- Позитивные кейсы: нарушение ловится ---

func (j *Job) CreateBatchJobs(in []CreateJob) error { return nil } // want `GID-102: method "CreateBatchJobs" contains the word Batch`

func (j *Job) BatchCreate(in []CreateJob) error { return nil } // want `GID-102: method "BatchCreate" contains the word Batch`

// Граничный кейс: Batch в середине имени.
func (j *Job) UpdateBatchStatus(status string) error { return nil } // want `GID-102: method "UpdateBatchStatus" contains the word Batch`

// --- Негативные кейсы: чистый код проходит ---

func (j *Job) CreateJob(in *CreateJob) error { return nil }

func (j *Job) CreateJobs(in []CreateJob) error { return nil }

// Граничный кейс: batch со строчной буквы — другое слово, не Batch-нейминг.
func (j *Job) processbatch() {}

// --- Неприменимость: правило действует только на методы ---

func BatchSize() int { return 100 }
