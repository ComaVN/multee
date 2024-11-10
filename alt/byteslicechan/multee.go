// Copyright 2023-2024 Roel Harbers.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Alternative implementation of the multee package, using channels.
// Implements a multiplexer for io.Readers, making it possible to read from a single io.Reader several times,
// without needing to Seek back to the beginning.
package multee

import (
	"io"
	"sync"
)

const bufferSize = 4096

type multeeReader struct {
	inputReader    io.Reader
	err            error
	InitReaderOnce *sync.Once           // This makes sure only a single reader will start the input reader.
	readers        map[*reader]struct{} // Unordered set of readers.
}

func NewMulteeReader(inputReader io.Reader) *multeeReader {
	return &multeeReader{
		inputReader:    inputReader,
		InitReaderOnce: new(sync.Once),
		readers:        make(map[*reader]struct{}),
	}
}

// Returns an io.ReadCloser. The caller must either keep reading until EOF or call Close(),
// or the MulteeReader will block.
// The returned reader is *not* concurrency-safe.
// Of course, it *is* safe to use multiple readers from the same multeeReader in different goroutines.
// TODO: at the moment, this method has not been checked for concurrency-safety, and it most likely isn't.
func (mr *multeeReader) NewReader() *reader {
	r := &reader{
		multeeReader: mr,
		c:            make(chan []byte, 1),
		closedC:      make(chan struct{}),
	}
	mr.readers[r] = struct{}{}
	return r
}

// Starts the goroutine to read from the input reader and multiplex to all reader channels.
func (mr *multeeReader) InitReader(wipr *reader) {
	go func() {
		// Loop while there is input, no errors, and unclosed readers.
		for func() bool {
			buf := make([]byte, bufferSize)
			if len(mr.readers) == 0 {
				return false
			}
			closedReaders := []*reader{}
			n, err := mr.inputReader.Read(buf)
			if n > 0 {
				for r := range mr.readers {
					select { // This blocks while the reader's channel is full and the reader is not closed.
					case r.c <- buf[:n]: // Send the current buffer to the reader's input channel.
					case <-r.closedC: // The reader has closed.
						closedReaders = append(closedReaders, r)
					}
				}
			}
			if err != nil {
				mr.err = err
				for r := range mr.readers {
					close(r.c)
				}
				return false
			}
			for _, r := range closedReaders {
				delete(mr.readers, r)
			}
			return true
		}() {
		}
	}()
}

// This is the io.ReadCloser returned by multiReaders.NewReader.
type reader struct {
	multeeReader *multeeReader
	c            chan []byte
	closed       bool
	closedC      chan struct{} // closing this channel signals to the multeeReader that the reader has closed.
	buf          []byte        // Buffer for misaligned reads.
}

func (r *reader) Read(p []byte) (n int, err error) {
	r.multeeReader.InitReaderOnce.Do(func() { r.multeeReader.InitReader(r) })
	n = 0
	if len(r.buf) > 0 {
		copied := copy(p, r.buf)
		n += copied
		if len(r.buf) >= copied {
			// p was completely filled by the buffer, buffer the rest (if any) for the next Read, and return.
			r.buf = r.buf[copied:]
			return n, nil
		}
	}
	for n < len(p) {
		bs, ok := <-r.c
		if ok {
			copied := copy(p[n:], bs)
			n += copied
			if len(bs) > copied {
				// Not all bytes from the channel's current byte slice fit into p, buffer the rest for the next Read.
				r.buf = bs[copied:]
			}
		} else {
			return n, r.multeeReader.err
		}
	}
	return n, nil
}

// TODO: at the moment, this method has not been checked for concurrency-safety, particularly with concurrent calls to newReader()
func (r *reader) Close() error {
	if r.closed {
		return ErrClosed
	}
	r.closed = true
	close(r.closedC)
	return nil
}
