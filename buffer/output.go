package buffer

// type that implements the output interface
type Output struct {
    buffer []byte
    offset int
}

// OutputBuffer is a buffer that can be written to.
// It is a wrapper around a byte slice.
// It implements the output interface.

func NewOutputBuffer(size int) *Output {
    return &Output{
        buffer: make([]byte, size),
        offset: 0,
    }
}

// Write writes the given byte slice to the buffer.
func (o *Output) Write(b []byte) {
    o.buffer = append(o.buffer, b...)
}
