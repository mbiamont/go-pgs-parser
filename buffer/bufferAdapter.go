package buffer

type BufferAdapter interface {
	Length() int
	At(index int) (int, error)
	SubArray(start int, end int) BufferAdapter
}
