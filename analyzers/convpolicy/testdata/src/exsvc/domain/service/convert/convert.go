package convert

// Source is the input enum.
type Source int

const (
	SourceFile    Source = 0
	SourceMeeting Source = 1
)

// asrFormatFromSource would be flagged, but is exempt via settings.exclude in
// the exclude test.
func asrFormatFromSource(source Source) uint32 {
	channels := uint32(1)
	if source == SourceMeeting {
		channels = 2
	}

	return channels
}

// SampleRateFromSource is NOT excluded — it still fires under the exclude test.
func SampleRateFromSource(source Source) uint32 {
	var rate uint32
	switch source {
	case SourceFile:
		rate = 8000 // want `GID-247: convert function "SampleRateFromSource" branches on input to select a constant value for "rate"`
	case SourceMeeting:
		rate = 48000
	}

	return rate
}
