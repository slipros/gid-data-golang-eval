package usecase

type Export struct{}

func (*Export) Work()    {}
func NewExport() *Export { return &Export{} }
