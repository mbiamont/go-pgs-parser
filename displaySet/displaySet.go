package displaySet

import (
	"errors"
	"github.com/mbiamont/go-pgs-parser/buffer"
	"github.com/mbiamont/go-pgs-parser/segment"
	"image"
	"image/color"
	"math"
	"time"
)

type ImageData struct {
	Image  image.Image
	Width  int
	Height int
}

type DisplaySet interface {
	paletteDefinitionSegment(paletteId int) (*segment.PaletteDefinitionSegment, error)

	parseImageData() (*ImageData, error)

	rleDecode(encodedBuffer buffer.BufferAdapter, callback func(int, int, int)) error

	ycrcbToRgba(palette segment.PaletteEntry) color.RGBA

	clamp(number float64, min int, max int) int

	objectData() []buffer.BufferAdapter

	paletteEntriesToRgba(entries []segment.PaletteEntry) []color.RGBA

	ToImageData() (*ImageData, error)

	StartTime() time.Duration
}

type displaySet struct {
	PresentationCompositionSegment segment.PresentationCompositionSegment
	WindowDefinitionSegments       []segment.WindowDefinitionSegment
	PaletteDefinitionSegments      []segment.PaletteDefinitionSegment
	ObjectDefinitionSegments       []segment.ObjectDefinitionSegment
	EndDefinitionSegment           segment.Segment

	PreviousDisplaySet *DisplaySet
}

func NewDisplaySet(
	presentationCompositionSegment segment.PresentationCompositionSegment,
	windowDefinitionSegments []segment.WindowDefinitionSegment,
	paletteDefinitionSegments []segment.PaletteDefinitionSegment,
	objectDefinitionSegments []segment.ObjectDefinitionSegment,
	endDefinitionSegment segment.Segment,
	previousDisplaySet *DisplaySet) DisplaySet {
	return &displaySet{
		PresentationCompositionSegment: presentationCompositionSegment,
		WindowDefinitionSegments:       windowDefinitionSegments,
		PaletteDefinitionSegments:      paletteDefinitionSegments,
		ObjectDefinitionSegments:       objectDefinitionSegments,
		EndDefinitionSegment:           endDefinitionSegment,
		PreviousDisplaySet:             previousDisplaySet,
	}
}

func (d *displaySet) ToImageData() (*ImageData, error) {
	if len(d.ObjectDefinitionSegments) <= 0 {
		//No image found
		return nil, nil
	}
	return d.parseImageData()
}

func (d *displaySet) firstOds() *segment.ObjectDefinitionSegment {
	for _, ods := range d.ObjectDefinitionSegments {
		if ods.LastInSequenceFlag == segment.LastInSequenceFlagFirstInSequence || ods.LastInSequenceFlag == segment.LastInSequenceFlagFirstAndLastInSequence {
			return &ods
		}
	}
	return nil
}

func (d *displaySet) paletteDefinitionSegment(paletteId int) (*segment.PaletteDefinitionSegment, error) {
	for _, pds := range d.PaletteDefinitionSegments {
		if pds.PaletteId == paletteId {
			return &pds, nil
		}
	}

	if d.PresentationCompositionSegment.CompositionState != segment.CompositionStateNormal {
		return nil, errors.New("PCS references invalid PDS and composition state is not 'normal'")
	}

	if d.PreviousDisplaySet != nil {
		return (*d.PreviousDisplaySet).paletteDefinitionSegment(paletteId)
	}

	return nil, errors.New("PCS references invalid PDS and no previous display set to fallback to")
}

func (d *displaySet) parseImageData() (*ImageData, error) {
	pds, err := d.paletteDefinitionSegment(d.PresentationCompositionSegment.PaletteId)

	if err != nil {
		return nil, err
	}

	if pds == nil {
		return nil, errors.New("PCS references invalid PDS")
	}

	firstOds := d.firstOds()

	if firstOds == nil || firstOds.Width == nil || firstOds.Height == nil {
		return nil, errors.New("missing first ODS with defined width and height")
	}
	rgbaPalette := d.paletteEntriesToRgba(pds.PaletteEntries)
	width := *firstOds.Width
	height := *firstOds.Height

	// New

	upLeft := image.Pt(0, 0)
	lowRight := image.Pt(width, height)
	img := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
	err = d.rleDecode(buffer.NewCompositeBuffer(d.objectData()), func(x int, y int, paletteIndex int) {

		if paletteIndex >= len(rgbaPalette) {
			img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		} else {
			rgba := rgbaPalette[paletteIndex]
			img.Set(x, y, rgba)
		}
	})

	if err != nil {
		return nil, err
	}

	return &ImageData{
		Image:  img,
		Width:  width,
		Height: height,
	}, nil
}

func (d *displaySet) StartTime() time.Duration {
	return d.EndDefinitionSegment.Header.StartTime
}

func (d *displaySet) rleDecode(encodedBuffer buffer.BufferAdapter, callback func(int, int, int)) error {
	encodedIndex := 0
	decodedLineIndex := 0
	currentLine := 0
	encodedLength := encodedBuffer.Length()

	for encodedIndex < encodedLength {
		firstByte, err := encodedBuffer.At(encodedIndex)

		if err != nil {
			return err
		}

		var runLength int
		var colorB int
		var increment int

		// Deal with each possible code
		if firstByte > 0 {
			// CCCCCCCC	- One pixel in color C
			colorB = firstByte
			runLength = 1
			increment = 1
		} else {
			secondByte, err := encodedBuffer.At(encodedIndex + 1)

			if err != nil {
				return err
			}

			if secondByte == 0 {
				// 00000000 00000000 - End of line
				colorB = 0
				runLength = 0
				increment = 2
				decodedLineIndex = 0
				currentLine++
			} else if secondByte < 64 {
				// 00000000 00LLLLLL - L pixels in color 0 (L between 1 and 63)
				colorB = 0
				runLength = secondByte
				increment = 2
			} else if secondByte < 128 {
				// 00000000 01LLLLLL LLLLLLLL - L pixels in color 0 (L between 64 and 16383)
				thirdByte, err := encodedBuffer.At(encodedIndex + 2)

				if err != nil {
					return err
				}

				colorB = 0
				runLength = ((secondByte - 64) << 8) + thirdByte
				increment = 3
			} else if secondByte < 192 {
				// 00000000 10LLLLLL CCCCCCCC - L pixels in color C (L between 3 and 63)
				thirdByte, err := encodedBuffer.At(encodedIndex + 2)

				if err != nil {
					return err
				}

				colorB = thirdByte
				runLength = secondByte - 128
				increment = 3
			} else {
				// 00000000 11LLLLLL LLLLLLLL CCCCCCCC - L pixels in color C (L between 64 and 16383)
				thirdByte, err := encodedBuffer.At(encodedIndex + 2)

				if err != nil {
					return err
				}

				fourthByte, err := encodedBuffer.At(encodedIndex + 3)

				if err != nil {
					return err
				}

				colorB = fourthByte
				runLength = ((secondByte - 192) << 8) + thirdByte
				increment = 4
			}
		}

		if runLength > 0 {
			for x := decodedLineIndex; x < decodedLineIndex+runLength; x++ {
				callback(x, currentLine, colorB)
			}

			decodedLineIndex += runLength
		}

		encodedIndex += increment
	}

	return nil
}

func (d *displaySet) ycrcbToRgba(palette segment.PaletteEntry) color.RGBA {
	y := float64(palette.Luminance)
	cb := float64(palette.ColorDifferenceBlue)
	cr := float64(palette.ColorDifferenceRed)

	r := d.clamp(math.Floor(y+1.4075*(cr-128)), 0, 255)
	g := d.clamp(math.Floor(y-0.3455*(cb-128)-0.7169*(cr-128)), 0, 255)
	b := d.clamp(math.Floor(y+1.779*(cb-128)), 0, 255)

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(palette.Transparency),
	}
}

func (d *displaySet) clamp(number float64, min int, max int) int {
	return int(math.Max(float64(min), math.Min(float64(max), number)))
}

func (d *displaySet) objectData() []buffer.BufferAdapter {
	var objectData []buffer.BufferAdapter

	for _, ods := range d.ObjectDefinitionSegments {
		objectData = append(objectData, ods.ObjectData)
	}

	return objectData
}

func (d *displaySet) paletteEntriesToRgba(entries []segment.PaletteEntry) []color.RGBA {
	var rgbas []color.RGBA

	for _, palette := range entries {
		rgbas = append(rgbas, d.ycrcbToRgba(palette))
	}

	return rgbas
}
