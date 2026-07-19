package service

// Source mirrors the convert enum.
type Source int

const (
	SourceFile    Source = 0
	SourceMeeting Source = 1
)

// pickChannels contains the exact policy-selection pattern, but lives outside a
// convert package — non-applicability: the rule targets convert packages only.
func pickChannels(source Source) uint32 {
	channels := uint32(1)
	if source == SourceMeeting {
		channels = 2
	}

	return channels
}
