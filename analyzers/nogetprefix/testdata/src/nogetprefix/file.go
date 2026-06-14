// Eval for GID-101 (no-get-prefix).
package nogetprefix

type Job struct {
	id     string
	status string
}

// --- Positive cases: the violation is caught ---

func (j *Job) GetID() string { return j.id } // want `GID-101: method "GetID" uses the Get prefix\. Fix: name getters without it: "ID"`

func (j *Job) GetStatus() string { return j.status } // want `GID-101: method "GetStatus" uses the Get prefix`

// Boundary case: a bare Get is also a violation.
func (j *Job) Get() string { return j.id } // want `GID-101: method "Get" uses the Get prefix`

// --- Negative cases: clean code passes ---

func (j *Job) ID() string { return j.id }

func (j *Job) Status() string { return j.status }

// Boundary case: Get is part of a word, not a prefix.
func (j *Job) Getaway() string { return j.id }

// --- Not applicable: the rule applies only to methods ---

func GetEnv(key string) string { return key }
