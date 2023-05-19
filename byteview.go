package ocache

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	bytes []byte
}

// Len returns the view's length
func (view ByteView) Len() int {
	return len(view.bytes)
}

// ByteSlice returns a copy of the data as a byte slice.
func (view ByteView) ByteSlice() []byte {
	return cloneBytes(view.bytes)
}

// String returns the data as a string, making a copy if necessary.
func (view ByteView) String() string {
	return string(view.bytes)
}

func cloneBytes(bytes []byte) []byte {
	newBytes := make([]byte, len(bytes))
	copy(newBytes, bytes)
	return newBytes
}
