// Package readers provides useful io.Reader related tools
package readers

import (
	"io"
)

// SizeableReader is a reader with a size
type SizeableReader interface {
	io.Reader
	Size() int64
}

// AtEOF reports whether the Reader is at EOF
type AtEOF interface {
	AtEOF() bool
}

// LimitedReadCloser adds io.Closer to io.LimitedReader.
type LimitedReadCloser struct {
	io.LimitedReader
	rc io.ReadCloser
}

// NewLimitedReadCloser creates an io.ReadCloser with a limit of n bytes.
// This is similar to io.LimitedReader but also provides Close functionality.
func NewLimitedReadCloser(rc io.ReadCloser, n int64) io.ReadCloser {
	return &LimitedReadCloser{
		LimitedReader: io.LimitedReader{
			R: rc,
			N: n,
		},
		rc: rc,
	}
}

// Close closes the underlying io.ReadCloser
func (l *LimitedReadCloser) Close() error {
	return l.rc.Close()
}

// ReadSeeker makes an io.ReadSeeker from an io.Reader by reading ahead
type ReadSeeker struct {
	r       io.Reader
	buf     []byte
	bufPos  int
	bufSize int
	offset  int64
}

// NewReadSeeker makes an io.ReadSeeker from an io.Reader
func NewReadSeeker(r io.Reader) io.ReadSeeker {
	return &ReadSeeker{
		r:   r,
		buf: make([]byte, 1024),
	}
}

// Read implements io.Reader
func (r *ReadSeeker) Read(p []byte) (n int, err error) {
	if r.bufPos < r.bufSize {
		n = copy(p, r.buf[r.bufPos:r.bufSize])
		r.bufPos += n
		r.offset += int64(n)
		if n > 0 {
			return n, nil
		}
	}
	n, err = r.r.Read(p)
	r.offset += int64(n)
	return n, err
}

// Seek implements io.Seeker
func (r *ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = r.offset + offset
	case io.SeekEnd:
		return 0, io.ErrUnexpectedEOF
	}
	if newOffset < 0 {
		return 0, io.ErrUnexpectedEOF
	}
	if newOffset == r.offset {
		return newOffset, nil
	}
	if newOffset < r.offset && newOffset >= r.offset-int64(r.bufSize) {
		// Seeking backwards into the buffer
		r.bufPos = int(int64(r.bufSize) - (r.offset - newOffset))
		r.offset = newOffset
		return newOffset, nil
	}
	return 0, io.ErrUnexpectedEOF
}

// StdoutLogger send bytes to a callback
type StdoutLogger struct {
	Callback func([]byte)
}

// Write logs the data sent to stdout
func (l *StdoutLogger) Write(p []byte) (n int, err error) {
	l.Callback(p)
	return len(p), nil
}

// Counter counts the bytes read
type Counter struct {
	total int64
}

// BytesRead returns the number of bytes read
func (c *Counter) BytesRead() int64 {
	if c == nil {
		return 0
	}
	return c.total
}

// Read bytes from the reader and count them
func (c *Counter) Read(b []byte) (n int, err error) {
	n = len(b)
	if c != nil {
		c.total += int64(n)
	}
	return n, nil
}
