// Eval for GID-102 (no-batch-word).
package nobatch

type Job struct{ id string }

type CreateJob struct{ ID string }

// --- Positive cases: the violation is caught ---

func (j *Job) CreateBatchJobs(in []CreateJob) error { return nil } // want `GID-102: method "CreateBatchJobs" contains the word Batch`

func (j *Job) BatchCreate(in []CreateJob) error { return nil } // want `GID-102: method "BatchCreate" contains the word Batch`

// Boundary case: Batch in the middle of the name.
func (j *Job) UpdateBatchStatus(status string) error { return nil } // want `GID-102: method "UpdateBatchStatus" contains the word Batch`

// --- Negative cases: clean code passes ---

func (j *Job) CreateJob(in *CreateJob) error { return nil }

func (j *Job) CreateJobs(in []CreateJob) error { return nil }

// Boundary case: lowercase batch is a different word, not Batch naming.
func (j *Job) processbatch() {}

// --- Not applicable: the rule applies only to methods ---

func BatchSize() int { return 100 }
