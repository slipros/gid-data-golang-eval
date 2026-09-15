package service

type Report struct {
	source interface{ Read() []byte }
}

func (r *Report) Build() {
	type reportSink interface{ Write([]byte) }
	var _ reportSink
}

const reportLimit = 5

var reportDefault = Report{}
