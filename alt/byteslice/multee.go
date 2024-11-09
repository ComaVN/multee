// Copyright 2023 Roel Harbers.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Alternative implementation of the multee package, using a synchronized byte slice as a buffer.
// Implements a multiplexer for io.Readers, making it possible to read from a single io.Reader several times,
// without needing to Seek back to the beginning.
package multee

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

const bufferSize = 32 * 1024

type multeeReader struct {
	inputReader      io.Reader
	err              error
	buf              [bufferSize]byte
	bufEndPos        int
	bufReadOnce      *sync.Once     // This makes sure only a single reader will load the next buffer.
	bufReadWaitGroup sync.WaitGroup // This waits for all current readers to be finished with reading the current buffer (or closed).
	readerCnt        atomic.Int32
}

func NewMulteeReader(inputReader io.Reader) *multeeReader {
	return &multeeReader{
		inputReader: inputReader,
		bufReadOnce: new(sync.Once),
	}
}

// Returns an io.ReadCloser. The caller must either keep reading until EOF or call Close(),
// or the MulteeReader will block.
// The returned reader is *not* concurrency-safe.
// Of course, it *is* safe to use multiple readers from the same multeeReader in different goroutines.
// TODO: at the moment, adding new readers to an existing multeeReader that's being read from by other readers is *not* safe.
// This probably should either be made safe or, more likely, impossible.
func (mr *multeeReader) NewReader() *reader {
	mr.readerCnt.Add(1)
	mr.bufReadWaitGroup.Add(1)
	return &reader{
		multeeReader: mr,
	}
}

// Used internally by reader to read buffered input bytes while keeping track of position.
// Returns the new buffer offset, the number of bytes read, and an error, if any.
func (mr *multeeReader) read(p []byte, bufOffset int) (int, int, error) {
	if bufOffset == mr.bufEndPos {
		// The current buffer is empty, or has been fully read by the calling reader.
		once := mr.bufReadOnce // This needs to be first, to prevent a race condition when the new "once" is created.
		mr.bufReadWaitGroup.Done()
		once.Do(func() {
			// Let the first reader to get here wait for the other readers, and buffer the next input block.
			// This means no-one else is accessing mr.buf, mr.bufEndPos, mr.inputReader or mr.err.
			// This is concurrency-safe as long as no readers are added while waiting here.
			mr.bufReadWaitGroup.Wait()
			mr.bufReadWaitGroup.Add(int(mr.readerCnt.Load()))
			mr.bufEndPos, mr.err = mr.inputReader.Read(mr.buf[:])
			mr.bufReadOnce = new(sync.Once)
		})
		bufOffset = 0
	}
	if bufOffset > mr.bufEndPos {
		// RH: ATTN: This should be impossible.
		panic(fmt.Errorf("reader buffer offset (%d) is beyond buffer end (%d)", bufOffset, mr.bufEndPos))
	}
	// Copy the remaining part of the buffer, or the size of p, whichever is smaller
	copied := copy(p, mr.buf[bufOffset:mr.bufEndPos])
	return bufOffset + copied, copied, mr.err
}

// This is the io.ReadCloser returned by multiReaders.NewReader
type reader struct {
	multeeReader *multeeReader
	bufOffset    int
	closed       bool
}

func (r *reader) Read(p []byte) (n int, err error) {
	r.bufOffset, n, err = r.multeeReader.read(p, r.bufOffset)
	return n, err
}

func (r *reader) Close() error {
	if r.closed {
		return ErrClosed
	}
	// readerCnt must be decremented before the wait group, because the readerCnt is used to increment the wait group again.
	r.multeeReader.readerCnt.Add(-1)
	r.multeeReader.bufReadWaitGroup.Done()
	r.closed = true
	return nil
}
