package buffer

import (
	"errors"
)

type CompositeBufferReader interface {
	Add(buffer []byte)

	Length() int

	ReadBytes(count int) (BufferAdapter, error)
}

type compositeBufferReader struct {
	buffers []BufferAdapter
}

func NewCompositeBufferReader() CompositeBufferReader {
	return &compositeBufferReader{}
}

func (c *compositeBufferReader) Add(buffer []byte) {
	c.buffers = append(c.buffers, NewUint8ArrayBuffer(buffer))
}

func (c *compositeBufferReader) Length() int {
	length := 0
	for _, b := range c.buffers {
		length += b.Length()
	}
	return length
}

func (c *compositeBufferReader) ReadBytes(count int) (BufferAdapter, error) {
	if count == 0 {
		return NewCompositeBuffer([]BufferAdapter{}), nil
	}

	var chunks []BufferAdapter
	accumulated := 0

	for true {
		if len(c.buffers) == 0 {
			return nil, errors.New("trying to read more bytes than available")
		}
		buffer := c.buffers[0]
		c.buffers = append(c.buffers[:0], c.buffers[1:]...)
		required := count - accumulated

		if buffer.Length() == required {
			chunks = append(chunks, buffer)
			break
		} else if buffer.Length() > required {
			chunks = append(chunks, buffer.SubArray(0, required))
			c.buffers = append([]BufferAdapter{buffer.SubArray(required, buffer.Length())}, c.buffers...)
			break
		}

		accumulated += buffer.Length()
		chunks = append(chunks, buffer)
	}

	return NewCompositeBuffer(chunks), nil
}
