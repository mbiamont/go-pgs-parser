package segment

import (
	"github.com/mbiamont/go-pgs-parser/buffer"
	"time"
)

type SegmentType uint8

const (
	SegmentTypePds SegmentType = iota
	SegmentTypeOds
	SegmentTypePcs
	SegmentTypeWds
	SegmentTypeEnd
)

type CompositionState uint8

const (
	CompositionStateNormal CompositionState = iota
	CompositionStateAcquisitionState
	CompositionStateEpochStart
)

type LastInSequenceFlag uint8

const (
	LastInSequenceFlagLastInSequence LastInSequenceFlag = iota
	LastInSequenceFlagFirstInSequence
	LastInSequenceFlagFirstAndLastInSequence
)

type SegmentHeader struct {
	PresentationTimestamp int
	DecodingTimestamp     int
	SegmentType           SegmentType
	SegmentSize           int
	StartTime             time.Duration
}

type Segment struct {
	Header SegmentHeader
}

type PresentationCompositionSegment struct {
	Width                            int
	Height                           int
	CompositionNumber                int
	CompositionState                 CompositionState
	PaletteUpdateFlag                bool
	PaletteId                        int
	CompositionObjectCount           int
	ObjectId                         int
	WindowId                         int
	ObjectCroppedFlag                bool
	ObjectHorizontalPosition         int
	ObjectVerticalPosition           int
	ObjectCroppingHorizontalPosition int
	ObjectCroppingVerticalPosition   int
	ObjectCroppingWidth              int
	ObjectCroppingHeight             int

	Segment
}

type WindowDefinition struct {
	WindowId                 int
	WindowHorizontalPosition int
	WindowVerticalPosition   int
	WindowWidth              int
	WindowHeight             int
}

type WindowDefinitionSegment struct {
	WindowCount       int
	WindowDefinitions []WindowDefinition

	Segment
}

type PaletteEntry struct {
	PaletteEntryId      int
	Luminance           int
	ColorDifferenceRed  int
	ColorDifferenceBlue int
	Transparency        int
}

type PaletteDefinitionSegment struct {
	PaletteId            int
	PaletteVersionNumber int
	PaletteEntries       []PaletteEntry

	Segment
}

type ObjectDefinitionSegment struct {
	ObjectId            int
	ObjectVersionNumber int
	LastInSequenceFlag  LastInSequenceFlag
	ObjectDataLength    int
	Width               *int
	Height              *int
	ObjectData          buffer.BufferAdapter

	Segment
}
