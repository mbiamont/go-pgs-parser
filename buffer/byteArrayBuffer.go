package buffer

import "errors"

type ByteArrayBuffer struct {
	buffer []byte
	BufferAdapter
}

func NewUint8ArrayBuffer(buffer []byte) *ByteArrayBuffer {
	return &ByteArrayBuffer{
		buffer: buffer,
	}
}

func (u *ByteArrayBuffer) Length() int {
	return len(u.buffer)
}

func (u *ByteArrayBuffer) At(index int) (int, error) {
	if index < len(u.buffer) {
		return int(u.buffer[index]), nil
	}

	return 0, errors.New("index out of bounds")
}

func (u *ByteArrayBuffer) SubArray(start int, end int) BufferAdapter {
	return NewUint8ArrayBuffer(u.buffer[start:end])
}
