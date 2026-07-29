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

// Boundary: every constructor of an entity — the exported one, an unexported
// factory, a second constructor — sits under the type declaration, above the
// methods; none of them splits the block.

type Worker struct {
	name string
}

func NewWorker(name string) *Worker { return &Worker{name: name} }

func newDefaultWorker() *Worker { return &Worker{name: "default"} }

func NewWorkerByTask(t Task) (*Worker, error) { return &Worker{name: t.name}, nil }

func (w *Worker) Name() string { return w.name }

func (w *Worker) Rename(name string) { w.name = name }

// Boundary: a bare New — the library idiom — is a constructor too, its entity
// read from the result type.

type Pool struct {
	size int
}

func New(size int) *Pool { return &Pool{size: size} }

func (p *Pool) Size() int { return p.size }
