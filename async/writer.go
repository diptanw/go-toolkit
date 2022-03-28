package async

import (
	"io"
	"sync"
)

// ConcurrentWriter is the io.Writer wrapper that enables concurrent use. Useful
// when parallel writing to io socket.
type ConcurrentWriter struct {
	io.Writer
	writeMu sync.Mutex
}

// Write writes to underlying io.Writer, guarding it with Mutex.
func (w *ConcurrentWriter) Write(b []byte) (int, error) {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	return w.Writer.Write(b)
}
