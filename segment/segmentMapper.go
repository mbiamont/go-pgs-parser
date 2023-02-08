package segment

import (
	"errors"
	"fmt"
)

type SegmentMapper interface {
	ToSegmentType(b byte) (SegmentType, error)

	ToCompositionState(b byte) (CompositionState, error)

	ToPaletteUpdateFlag(b byte) (bool, error)

	ToObjectCroppedFlag(b byte) (bool, error)

	ToLastInSequenceFlag(b byte) (LastInSequenceFlag, error)
}

func NewSegmentMapper() SegmentMapper {
	return &segmentMapper{}
}

type segmentMapper struct {
}

func (*segmentMapper) ToSegmentType(b byte) (SegmentType, error) {
	switch b {
	case 20:
		return SegmentTypePds, nil
	case 21:
		return SegmentTypeOds, nil
	case 22:
		return SegmentTypePcs, nil
	case 23:
		return SegmentTypeWds, nil
	case 128:
		return SegmentTypeEnd, nil
	}

	return 0, errors.New(fmt.Sprintf("invalid segment type byte: %x", b))
}

func (*segmentMapper) ToCompositionState(b byte) (CompositionState, error) {
	switch b {
	case 0:
		return CompositionStateNormal, nil
	case 64:
		return CompositionStateAcquisitionState, nil
	case 128:
		return CompositionStateEpochStart, nil
	}

	return 0, errors.New(fmt.Sprintf("invalid composition state byte: %x", b))
}

func (*segmentMapper) ToPaletteUpdateFlag(b byte) (bool, error) {
	switch b {
	case 0:
		return false, nil
	case 128:
		return true, nil
	}

	return false, errors.New(fmt.Sprintf("invalid palette update flag byte: %x", b))
}

func (*segmentMapper) ToObjectCroppedFlag(b byte) (bool, error) {
	switch b {
	case 0:
		return false, nil
	case 64:
		return true, nil
	}

	return false, errors.New(fmt.Sprintf("invalid object cropped flag byte: %x", b))
}

func (*segmentMapper) ToLastInSequenceFlag(b byte) (LastInSequenceFlag, error) {
	switch b {
	case 64:
		return LastInSequenceFlagLastInSequence, nil
	case 128:
		return LastInSequenceFlagFirstInSequence, nil
	case 192:
		return LastInSequenceFlagFirstAndLastInSequence, nil
	}

	return 0, errors.New(fmt.Sprintf("invalid object cropped flag byte: %x", b))
}
