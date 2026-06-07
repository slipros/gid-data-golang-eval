package entity

type Snapshot struct {
	ID   string
	Name string
}

type CreateSnapshot struct {
	Name string
}

type Snapshots []Snapshot
