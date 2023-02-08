package buffer

import (
	"errors"
	"math"
)

type CompositeBuffer struct {
	buffers []BufferAdapter
}

func NewCompositeBuffer(buffers []BufferAdapter) *CompositeBuffer {
	return &CompositeBuffer{
		buffers: buffers,
	}
}

func (c *CompositeBuffer) Length() int {
	length := 0
	for _, b := range c.buffers {
		length += b.Length()
	}
	return length
}

func (c *CompositeBuffer) At(index int) (int, error) {
	previousBuffersLength := 0

	for _, buffer := range c.buffers {
		bufferIndex := index - previousBuffersLength

		if bufferIndex < buffer.Length() {
			return buffer.At(bufferIndex)
		}

		previousBuffersLength += buffer.Length()
	}

	return 0, errors.New("index out of bounds")
}

func (c *CompositeBuffer) SubArray(start int, end int) BufferAdapter {
	var chunks []BufferAdapter
	previousBuffersLength := 0

	for _, buffer := range c.buffers {
		startBufferIndex := int(math.Max(0, float64(start-previousBuffersLength)))
		endBufferIndex := int(math.Min(float64(buffer.Length()), float64(end-previousBuffersLength)))

		if endBufferIndex > 0 && startBufferIndex < endBufferIndex {
			chunks = append(chunks, buffer.SubArray(startBufferIndex, endBufferIndex))
		}

		previousBuffersLength += buffer.Length()
	}

	return NewCompositeBuffer(chunks)
}
