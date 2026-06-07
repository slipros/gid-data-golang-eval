// Позитив: метод и конструктор сущности в чужом файле.
package entitygroup

func NewJob() *Job { // want `GID-157: "NewJob" принадлежит сущности "Job" — код сущности живёт в файле её объявления`
	return &Job{}
}

func (s *Snapshot) Foreign() string { // want `GID-157: "Foreign" принадлежит сущности "Snapshot" — код сущности живёт в файле её объявления`
	return s.name
}
