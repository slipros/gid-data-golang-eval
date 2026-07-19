package convert

// Source is the input enum the converters branch on.
type Source int

const (
	SourceFile    Source = 0
	SourceMeeting Source = 1
)

// AudioFormat is the destination vocabulary struct.
type AudioFormat struct {
	Codec      string
	SampleRate uint32
	Channels   uint32
}

// Codec is a named enum type — a legitimate mapping target.
type Codec string

const (
	CodecOpus Codec = "opus"
	CodecAAC  Codec = "aac"
)

// --- positive: policy selection of a raw value by input ---

// AudioFormatFromSource invents an audio format and picks the channel count by
// input — business policy, not mapping.
func AudioFormatFromSource(source Source) AudioFormat {
	const (
		codec     = "opus"
		rate      = 48000
		chMeeting = 2
		chFile    = 1
	)

	channels := uint32(chFile)
	if source == SourceMeeting {
		channels = chMeeting // want `GID-247: convert function "AudioFormatFromSource" branches on input to select a constant value for "channels"`
	}

	return AudioFormat{Codec: codec, SampleRate: rate, Channels: channels}
}

// SampleRateFromSource selects a raw sample rate via a switch on input.
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

// --- negative: pure mapping ---

// AudioFormatCopy copies fields from the input — no invented values.
func AudioFormatCopy(in AudioFormat) AudioFormat {
	out := AudioFormat{}
	out.Codec = in.Codec
	out.SampleRate = in.SampleRate
	out.Channels = in.Channels

	return out
}

// CodecFromSource maps input to a named enum type — vocabulary mapping, not a
// raw policy value (GID-143/233 territory).
func CodecFromSource(source Source) Codec {
	var c Codec
	switch source {
	case SourceMeeting:
		c = CodecOpus
	case SourceFile:
		c = CodecAAC
	}

	return c
}

// --- boundary: branch not on input ---

// ChannelsNormalized branches on a local, not on an input parameter — the
// decision is derived from the copied value, not policy over the input.
func ChannelsNormalized(in AudioFormat) uint32 {
	channels := in.Channels
	if channels == 0 {
		channels = 1
	}

	return channels
}

// LocalBranch branches on a locally computed value, not the parameter.
func LocalBranch(in AudioFormat) uint32 {
	n := in.Channels
	x := uint32(1)
	if n > 2 {
		x = 3
	}

	return x
}

// --- boundary: same constant in every branch (no selection) ---

// SameConst assigns the same value regardless of input — a single distinct
// value, so not a selection.
func SameConst(source Source) uint32 {
	var x uint32
	if source == SourceMeeting {
		x = 5
	} else {
		x = 5
	}

	return x
}
