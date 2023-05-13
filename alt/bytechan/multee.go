// Copyright 2023-2024 Roel Harbers.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Alternative implementation of the multee package, using channels.
// This is very slow, mostly because we send every individual byte on all channels.
// Implements a multiplexer for io.Readers, making it possible to read from a single io.Reader several times,
// without needing to Seek back to the beginning.
package bytechan

import (
	"io"
	"sync"

	"github.com/ComaVN/multee"
)

const bufferSize = 1024

type readers struct {
	sync.Mutex
	m map[*reader]struct{}
}

type multeeReader struct {
	inputReader io.Reader
	err         error
	bufReadOnce *sync.Once // This makes sure only a single reader will load the next buffer.
	readers     readers
}

func NewMulteeReader(inputReader io.Reader) *multeeReader {
	return &multeeReader{
		inputReader: inputReader,
		bufReadOnce: new(sync.Once),
		readers:     readers{m: map[*reader]struct{}{}},
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
		c:            make(chan byte, bufferSize),
	}
	mr.readers.Lock()
	defer mr.readers.Unlock()
	mr.readers.m[r] = struct{}{}
	return r
}

func (mr *multeeReader) readAndSendToAllReaders() {
	mr.bufReadOnce.Do(func() {
		buf := make([]byte, bufferSize)
		n, err := mr.inputReader.Read(buf)
		mr.readers.Lock()
		defer mr.readers.Unlock()
		for _, b := range buf[:n] {
			for r := range mr.readers.m {
				r.c <- b
			}
		}
		if err != nil {
			mr.err = err
			for r := range mr.readers.m {
				close(r.c)
			}
		}
		mr.bufReadOnce = new(sync.Once)
	})
}

// This is the io.ReadCloser returned by multiReaders.NewReader
type reader struct {
	multeeReader *multeeReader
	c            chan byte
	closed       bool
}

func (r *reader) Read(p []byte) (n int, err error) {
	n = 0
	readNext := true
	bytesAvailable := true
	for n < len(p) && (bytesAvailable || readNext) {
		select {
		case b, ok := <-r.c:
			if ok {
				p[n] = b
				n++
			} else {
				return n, r.multeeReader.err
			}
		default:
			if readNext {
				readNext = false
				r.multeeReader.readAndSendToAllReaders()
			} else {
				bytesAvailable = false
			}
		}
	}
	return n, nil
}

// TODO: at the moment, this method has not been checked for concurrency-safety, particularly with concurrent calls to newReader()
func (r *reader) Close() error {
	if r.closed {
		return multee.ErrClosed
	}
	r.closed = true
	r.multeeReader.readers.Lock()
	defer r.multeeReader.readers.Unlock()
	delete(r.multeeReader.readers.m, r)
	return nil
}
