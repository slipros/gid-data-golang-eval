package entity

type Snapshot struct {
	ID string
}

type CreateSnapshot struct {
	Name string
}

type Snapshots []Snapshot
