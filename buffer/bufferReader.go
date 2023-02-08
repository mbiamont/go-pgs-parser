package buffer

type BufferReader interface {
	Index() int
	HasNext() bool
	ReadBytes(count int) (int, error)
	ReadBytesWithLimit(count int, limit *int) (int, error)
	ReadBuffer(count int) BufferAdapter
}

type bufferReader struct {
	buffer BufferAdapter
	index  int
}

func NewBufferReader(bufferAdapter BufferAdapter) BufferReader {
	return &bufferReader{
		buffer: bufferAdapter,
		index:  0,
	}
}

func (b *bufferReader) Index() int {
	return b.index
}

func (b *bufferReader) HasNext() bool {
	return b.index < b.buffer.Length()
}

func (b *bufferReader) ReadBytes(count int) (int, error) {
	return b.ReadBytesWithLimit(count, nil)
}

func (b *bufferReader) ReadBytesWithLimit(count int, limit *int) (int, error) {
	if limit != nil && b.index+count > *limit {
		return 0, nil
	}

	number := 0
	digit := 0
	from := b.index
	to := b.index + count - 1

	for i := to; i >= from; i-- {
		bb, err := b.buffer.At(i)

		if err != nil {
			return 0, err
		}
		number += bb << (8 * digit)
		digit++
	}

	b.index += count

	return number, nil
}

func (b *bufferReader) ReadBuffer(count int) BufferAdapter {
	buffer := b.buffer.SubArray(b.index, b.index+count)
	b.index += count

	return buffer
}
