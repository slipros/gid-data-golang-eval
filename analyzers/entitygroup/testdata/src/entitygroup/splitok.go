// Negative: free declarations outside the entity blocks — either edge is fine.
package entitygroup

// Not applicable: a free helper above the first type.
func normalize(s string) string { return s }

type Task struct {
	name string
}

func (t *Task) Name() string { return t.name }

// Boundary: a free function exactly between two entity blocks — it splits nothing.
func taskKey(name string) string { return "task:" + normalize(name) }

type Queue struct {
	tasks []Task
}

func (q *Queue) Len() int { return len(q.tasks) }

// Not applicable: free helpers below the last method of the last entity.
func drain(q *Queue) { q.tasks = nil }
