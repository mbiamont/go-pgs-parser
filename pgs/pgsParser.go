package pgs

import (
	"github.com/mbiamont/go-pgs-parser/buffer"
	"github.com/mbiamont/go-pgs-parser/displaySet"
	"image/jpeg"
	"image/png"
	"os"
	"time"
)

type PgsParser interface {
	// ParsePgsFile Parse the input file path and call the onImage function for each ImageData found
	ParsePgsFile(inputFilePath string, onImage func(index int, startTime time.Duration, data displaySet.ImageData) error) error

	// ParseDisplaySets Parse the input file path and call the onDisplaySet function for each DisplaySet found
	ParseDisplaySets(inputFilePath string, onDisplaySet func(data displaySet.DisplaySet, startTime time.Duration) error) error

	// ConvertToPngImages Parse the input file path and save each subtitle picture as a PNG using fileCreator function to create the PNG file
	ConvertToPngImages(inputFilePath string, fileCreator func(index int, startTime time.Duration) (*os.File, error)) error

	// ConvertToJpgImages Parse the input file path and save each subtitle picture as a JPG using fileCreator function to create the JPG file
	ConvertToJpgImages(inputFilePath string, fileCreator func(index int, startTime time.Duration) (*os.File, error)) error
}

type pgsParser struct {
}

// NewPgsParser Initialize a new PGS parser
func NewPgsParser() PgsParser {
	return &pgsParser{}
}

func (p *pgsParser) ParsePgsFile(inputFilePath string, onImage func(index int, startTime time.Duration, data displaySet.ImageData) error) error {
	i := 0
	return p.ParseDisplaySets(inputFilePath, func(data displaySet.DisplaySet, startTime time.Duration) error {
		imageData, err := data.ToImageData()

		if err != nil {
			return err
		} else if imageData != nil {
			err = onImage(i, startTime, *imageData)

			if err != nil {
				return err
			}
			i++
		}
		return nil
	})
}

func (p *pgsParser) ParseDisplaySets(inputFilePath string, onDisplaySet func(data displaySet.DisplaySet, startTime time.Duration) error) error {
	file, err := os.ReadFile(inputFilePath)

	if err != nil {
		return err
	}

	parser := displaySet.NewDisplaySetParser()
	accumulatedBuffer := buffer.NewCompositeBufferReader()
	requestedBytes := 13

	accumulatedBuffer.Add(file)

	for accumulatedBuffer.Length() >= requestedBytes {
		bytes, err := accumulatedBuffer.ReadBytes(requestedBytes)

		if err != nil {
			return err
		}

		requestedBytes, err = parser.Consume(bytes)

		if err != nil {
			return err
		}

		if parser.IsReady() {
			ds := parser.Next()
			if ds != nil {
				set := *ds
				err = onDisplaySet(set, set.StartTime())

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p *pgsParser) ConvertToPngImages(inputFilePath string, fileCreator func(index int, startTime time.Duration) (*os.File, error)) error {
	return p.ParsePgsFile(inputFilePath, func(index int, startTime time.Duration, data displaySet.ImageData) error {
		f, err := fileCreator(index, startTime)

		if err != nil {
			return err
		}

		defer f.Close()

		if err != nil {
			return err
		}

		return png.Encode(f, data.Image)
	})
}

func (p *pgsParser) ConvertToJpgImages(inputFilePath string, fileCreator func(index int, startTime time.Duration) (*os.File, error)) error {
	return p.ParsePgsFile(inputFilePath, func(index int, startTime time.Duration, data displaySet.ImageData) error {
		f, err := fileCreator(index, startTime)

		if err != nil {
			return err
		}

		defer f.Close()

		if err != nil {
			return err
		}

		return jpeg.Encode(f, data.Image, &jpeg.Options{
			Quality: 100,
		})
	})
}
