package displaySet

import (
	"errors"
	"fmt"
	"github.com/mbiamont/go-pgs-parser/buffer"
	"github.com/mbiamont/go-pgs-parser/segment"
	"time"
)

const pgMagicNumber = 20551 // 0x5047

type DisplaySetParser interface {
	Next() *DisplaySet

	IsReady() bool

	Consume(buffer buffer.BufferAdapter) (int, error)

	ParsePcsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.PresentationCompositionSegment, error)

	ParseWdsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.WindowDefinitionSegment, error)

	ParsePdsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.PaletteDefinitionSegment, error)

	ParseOdsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.ObjectDefinitionSegment, error)
}

type displaySetParser struct {
	SegmentMapper segment.SegmentMapper

	Header                         *segment.SegmentHeader
	LastDisplaySet                 *DisplaySet
	PresentationCompositionSegment *segment.PresentationCompositionSegment
	WindowDefinitionSegments       []segment.WindowDefinitionSegment
	PaletteDefinitionSegments      []segment.PaletteDefinitionSegment
	ObjectDefinitionSegments       []segment.ObjectDefinitionSegment

	Ready bool
}

func NewDisplaySetParser() DisplaySetParser {
	return &displaySetParser{
		SegmentMapper:                  segment.NewSegmentMapper(),
		Header:                         nil,
		LastDisplaySet:                 nil,
		PresentationCompositionSegment: nil,
		WindowDefinitionSegments:       []segment.WindowDefinitionSegment{},
		PaletteDefinitionSegments:      []segment.PaletteDefinitionSegment{},
		ObjectDefinitionSegments:       []segment.ObjectDefinitionSegment{},
		Ready:                          false,
	}
}

func (d *displaySetParser) Next() *DisplaySet {
	d.Ready = false
	return d.LastDisplaySet
}

func (d *displaySetParser) IsReady() bool {
	return d.Ready
}

func (d *displaySetParser) Consume(bf buffer.BufferAdapter) (int, error) {
	reader := buffer.NewBufferReader(bf)

	if d.Header != nil {
		switch d.Header.SegmentType {
		case segment.SegmentTypePcs:
			if d.PresentationCompositionSegment != nil {
				return 0, errors.New("unexpected PDS segment")
			}
			pcs, err := d.ParsePcsSegment(reader, *d.Header)

			if err != nil {
				return 0, err
			}

			d.PresentationCompositionSegment = pcs
			break
		case segment.SegmentTypeWds:
			if d.WindowDefinitionSegments == nil {
				return 0, errors.New("unexpected WDS segment")
			}
			wds, err := d.ParseWdsSegment(reader, *d.Header)

			if err != nil {
				return 0, err
			}

			d.WindowDefinitionSegments = append(d.WindowDefinitionSegments, *wds)
			break
		case segment.SegmentTypePds:
			if d.PaletteDefinitionSegments == nil {
				return 0, errors.New("unexpected PDS segment")
			}
			pds, err := d.ParsePdsSegment(reader, *d.Header)

			if err != nil {
				return 0, err
			}

			d.PaletteDefinitionSegments = append(d.PaletteDefinitionSegments, *pds)
			break
		case segment.SegmentTypeOds:
			if d.ObjectDefinitionSegments == nil {
				return 0, errors.New("unexpected ODS segment")
			}
			ods, err := d.ParseOdsSegment(reader, *d.Header)

			if err != nil {
				return 0, err
			}

			d.ObjectDefinitionSegments = append(d.ObjectDefinitionSegments, *ods)
			break
		case segment.SegmentTypeEnd:
			if d.PresentationCompositionSegment == nil {
				return 0, errors.New("unexpected END segment")
			}

			endDefinitionSegment := segment.Segment{
				Header: *d.Header,
			}

			lastDisplaySet := NewDisplaySet(
				*d.PresentationCompositionSegment,
				d.WindowDefinitionSegments,
				d.PaletteDefinitionSegments,
				d.ObjectDefinitionSegments,
				endDefinitionSegment,
				d.LastDisplaySet,
			)

			d.LastDisplaySet = &lastDisplaySet

			d.Ready = true
			d.PresentationCompositionSegment = nil
			d.WindowDefinitionSegments = []segment.WindowDefinitionSegment{}
			d.PaletteDefinitionSegments = []segment.PaletteDefinitionSegment{}
			d.ObjectDefinitionSegments = []segment.ObjectDefinitionSegment{}
			break

		default:
			return 0, fmt.Errorf("unknown segment type %d", d.Header.SegmentType)
		}

		d.Header = nil
		return 13, nil
	} else {
		magicNumber, err := reader.ReadBytes(2)

		if err != nil {
			return 0, err
		}

		if magicNumber != pgMagicNumber {
			return 0, fmt.Errorf("invalid magic number %d", magicNumber)
		}

		presentationTimestamp, err := reader.ReadBytes(4)

		if err != nil {
			return 0, err
		}

		decodingTimestamp, err := reader.ReadBytes(4)

		if err != nil {
			return 0, err
		}

		segmentTypeByte, err := reader.ReadBytes(1)

		if err != nil {
			return 0, err
		}

		segmentType, err := d.SegmentMapper.ToSegmentType(byte(segmentTypeByte))

		if err != nil {
			return 0, err
		}

		segmentSize, err := reader.ReadBytes(2)

		if err != nil {
			return 0, err
		}

		startTimeStamp := time.Duration(presentationTimestamp/90) * time.Millisecond

		d.Header = &segment.SegmentHeader{
			PresentationTimestamp: presentationTimestamp,
			DecodingTimestamp:     decodingTimestamp,
			SegmentType:           segmentType,
			SegmentSize:           segmentSize,
			StartTime:             startTimeStamp,
		}

		return segmentSize, nil
	}
}

func (d *displaySetParser) ParsePcsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.PresentationCompositionSegment, error) {
	limit := reader.Index() + header.SegmentSize
	width, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}

	height, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}

	_, err = reader.ReadBytes(1) // ignore frame rate

	if err != nil {
		return nil, err
	}

	compositionNumber, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}

	compositionStateByte, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	compositionState, err := d.SegmentMapper.ToCompositionState(byte(compositionStateByte))

	if err != nil {
		return nil, err
	}

	paletteUpdateFlagByte, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	paletteUpdateFlag, err := d.SegmentMapper.ToPaletteUpdateFlag(byte(paletteUpdateFlagByte))

	if err != nil {
		return nil, err
	}

	paletteId, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	compositionObjectCount, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	objectId, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}

	windowId, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	objectCroppedFlagByte, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	objectCroppedFlag, err := d.SegmentMapper.ToObjectCroppedFlag(byte(objectCroppedFlagByte))

	if err != nil {
		return nil, err
	}

	objectHorizontalPosition, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}
	objectVerticalPosition, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}
	objectCroppingHorizontalPosition, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}
	objectCroppingVerticalPosition, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}
	objectCroppingWidth, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}
	objectCroppingHeightPosition, err := reader.ReadBytesWithLimit(2, &limit)

	if err != nil {
		return nil, err
	}

	return &segment.PresentationCompositionSegment{
		Width:                            width,
		Height:                           height,
		CompositionNumber:                compositionNumber,
		CompositionState:                 compositionState,
		PaletteUpdateFlag:                paletteUpdateFlag,
		PaletteId:                        paletteId,
		CompositionObjectCount:           compositionObjectCount,
		ObjectId:                         objectId,
		WindowId:                         windowId,
		ObjectCroppedFlag:                objectCroppedFlag,
		ObjectHorizontalPosition:         objectHorizontalPosition,
		ObjectVerticalPosition:           objectVerticalPosition,
		ObjectCroppingHorizontalPosition: objectCroppingHorizontalPosition,
		ObjectCroppingVerticalPosition:   objectCroppingVerticalPosition,
		ObjectCroppingWidth:              objectCroppingWidth,
		ObjectCroppingHeight:             objectCroppingHeightPosition,
		Segment: segment.Segment{
			Header: header,
		},
	}, nil
}

func (d *displaySetParser) ParseWdsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.WindowDefinitionSegment, error) {
	limit := reader.Index() + header.SegmentSize
	windowCount, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	var windowDefinitions []segment.WindowDefinition

	for i := 0; i < windowCount; i++ {
		windowId, err := reader.ReadBytesWithLimit(1, &limit)

		if err != nil {
			return nil, err
		}
		windowHorizontalPosition, err := reader.ReadBytesWithLimit(2, &limit)

		if err != nil {
			return nil, err
		}
		windowVerticalPosition, err := reader.ReadBytesWithLimit(2, &limit)

		if err != nil {
			return nil, err
		}
		windowWidth, err := reader.ReadBytesWithLimit(2, &limit)

		if err != nil {
			return nil, err
		}
		windowHeight, err := reader.ReadBytesWithLimit(2, &limit)

		if err != nil {
			return nil, err
		}

		windowDefinitions = append(windowDefinitions, segment.WindowDefinition{
			WindowId:                 windowId,
			WindowHorizontalPosition: windowHorizontalPosition,
			WindowVerticalPosition:   windowVerticalPosition,
			WindowWidth:              windowWidth,
			WindowHeight:             windowHeight,
		})
	}

	return &segment.WindowDefinitionSegment{
		WindowCount:       windowCount,
		WindowDefinitions: windowDefinitions,
		Segment: segment.Segment{
			Header: header,
		},
	}, nil
}

func (d *displaySetParser) ParsePdsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.PaletteDefinitionSegment, error) {
	limit := reader.Index() + header.SegmentSize
	paletteId, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	paletteVersionNumber, err := reader.ReadBytesWithLimit(1, &limit)

	if err != nil {
		return nil, err
	}

	var paletteEntries []segment.PaletteEntry

	for reader.Index() < limit {
		paletteEntryId, err := reader.ReadBytesWithLimit(1, &limit)
		if err != nil {
			return nil, err
		}
		luminance, err := reader.ReadBytesWithLimit(1, &limit)
		if err != nil {
			return nil, err
		}
		colorDifferenceRed, err := reader.ReadBytesWithLimit(1, &limit)
		if err != nil {
			return nil, err
		}
		colorDifferenceBlue, err := reader.ReadBytesWithLimit(1, &limit)
		if err != nil {
			return nil, err
		}
		transparency, err := reader.ReadBytesWithLimit(1, &limit)
		if err != nil {
			return nil, err
		}

		paletteEntries = append(paletteEntries, segment.PaletteEntry{
			PaletteEntryId:      paletteEntryId,
			Luminance:           luminance,
			ColorDifferenceRed:  colorDifferenceRed,
			ColorDifferenceBlue: colorDifferenceBlue,
			Transparency:        transparency,
		})
	}

	return &segment.PaletteDefinitionSegment{
		PaletteId:            paletteId,
		PaletteVersionNumber: paletteVersionNumber,
		PaletteEntries:       paletteEntries,

		Segment: segment.Segment{
			Header: header,
		},
	}, nil
}

func (d *displaySetParser) ParseOdsSegment(reader buffer.BufferReader, header segment.SegmentHeader) (*segment.ObjectDefinitionSegment, error) {
	objectId, err := reader.ReadBytes(2)

	if err != nil {
		return nil, err
	}

	objectVersionNumber, err := reader.ReadBytes(1)

	if err != nil {
		return nil, err
	}

	lastInSequenceFlagByte, err := reader.ReadBytes(1)

	if err != nil {
		return nil, err
	}

	lastInSequenceFlag, err := d.SegmentMapper.ToLastInSequenceFlag(byte(lastInSequenceFlagByte))

	if err != nil {
		return nil, err
	}

	objectDataLength, err := reader.ReadBytes(3)

	var width *int = nil
	var height *int = nil
	var objectData buffer.BufferAdapter

	if lastInSequenceFlag == segment.LastInSequenceFlagFirstInSequence || lastInSequenceFlag == segment.LastInSequenceFlagFirstAndLastInSequence {
		w, e := reader.ReadBytes(2)

		if e != nil {
			return nil, e
		}

		h, e := reader.ReadBytes(2)

		if e != nil {
			return nil, e
		}

		width = &w
		height = &h
		objectData = reader.ReadBuffer(objectDataLength - 4)
	} else {
		objectData = reader.ReadBuffer(objectDataLength)
	}

	return &segment.ObjectDefinitionSegment{
		ObjectId:            objectId,
		ObjectVersionNumber: objectVersionNumber,
		LastInSequenceFlag:  lastInSequenceFlag,
		ObjectDataLength:    objectDataLength,
		Width:               width,
		Height:              height,
		ObjectData:          objectData,

		Segment: segment.Segment{
			Header: header,
		},
	}, nil
}
