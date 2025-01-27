// Copyright 2023-2025 Roel Harbers.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package multee

import (
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMulteeReader(t *testing.T) {
	mr := NewMulteeReader(strings.NewReader("foo"))
	assert.Equal(t, 0, mr.bufEndPos)
	assert.Equal(t, bufferSize, len(mr.buf))
}

func Test_multeeReader_NewReader(t *testing.T) {
	mr := NewMulteeReader(strings.NewReader("foo"))
	r := mr.NewReader()
	assert.False(t, r.closed)
	assert.Equal(t, 0, r.bufOffset)
}

func Test_multeeReader_read_impossible_offset(t *testing.T) {
	mr := NewMulteeReader(strings.NewReader("foo"))
	r := mr.NewReader()
	defer r.Close()
	assert.Panics(t, func() {
		_, _, _ = mr.read(make([]byte, 3), 1)
	})
}

func Test_reader_Read(t *testing.T) {
	t.Run("Single_reader_empty_input", func(t *testing.T) {
		ir := strings.NewReader("")
		mr := NewMulteeReader(ir)
		r := mr.NewReader()
		p := make([]byte, 4)
		bytesRead, err := r.Read(p)
		assert.Equal(t, 0, bytesRead)
		assert.Equal(t, io.EOF, err)
	})
	t.Run("Single_reader_short_input_single_read_needed", func(t *testing.T) {
		ir := strings.NewReader("foo")
		mr := NewMulteeReader(ir)
		r := mr.NewReader()
		p := make([]byte, 4)
		bytesRead, err := r.Read(p)
		assert.Equal(t, 3, bytesRead)
		assert.Equal(t, []byte("foo"), p[0:bytesRead])
		if assert.Nil(t, err) {
			bytesRead, err = r.Read(p)
			assert.Equal(t, 0, bytesRead)
			assert.Equal(t, io.EOF, err)
		}
	})
	t.Run("Single_reader_short_input_single_read_needed_exactly", func(t *testing.T) {
		ir := strings.NewReader("foo")
		mr := NewMulteeReader(ir)
		r := mr.NewReader()
		p := make([]byte, 3)
		bytesRead, err := r.Read(p)
		assert.Equal(t, 3, bytesRead)
		assert.Equal(t, []byte("foo"), p[0:bytesRead])
		if assert.Nil(t, err) {
			bytesRead, err = r.Read(p)
			assert.Equal(t, 0, bytesRead)
			assert.Equal(t, io.EOF, err)
		}
	})
	t.Run("Single_reader_short_input_multiple_reads_needed", func(t *testing.T) {
		ir := strings.NewReader("foobar!")
		mr := NewMulteeReader(ir)
		r := mr.NewReader()
		p := make([]byte, 3)
		bytesRead, err := r.Read(p)
		assert.Equal(t, 3, bytesRead)
		assert.Equal(t, []byte("foo"), p[0:bytesRead])
		if assert.Nil(t, err) {
			bytesRead, err = r.Read(p)
			assert.Equal(t, 3, bytesRead)
			assert.Equal(t, []byte("bar"), p[0:bytesRead])
			if assert.Nil(t, err) {
				bytesRead, err = r.Read(p)
				assert.Equal(t, 1, bytesRead)
				assert.Equal(t, []byte("!"), p[0:bytesRead])
				if assert.Nil(t, err) {
					bytesRead, err = r.Read(p)
					assert.Equal(t, 0, bytesRead)
					assert.Equal(t, io.EOF, err)
				}
			}
		}
	})
	t.Run("Three_readers_short_input_one_closes", func(t *testing.T) {
		// Note that you should probably never use two readers from a single goroutine.
		// This only works here because only a single reader reaches the end of the buffer.
		ir := strings.NewReader("foobar")
		mr := NewMulteeReader(ir)
		r1 := mr.NewReader()
		r2 := mr.NewReader()
		r3 := mr.NewReader()
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			p := make([]byte, 4)
			bytesRead, err := r1.Read(p)
			assert.Equal(t, 4, bytesRead)
			assert.Equal(t, []byte("foob"), p[0:bytesRead])
			if assert.Nil(t, err) {
				bytesRead, err = r1.Read(p)
				assert.Equal(t, 2, bytesRead)
				assert.Equal(t, []byte("ar"), p[0:bytesRead])
				if assert.Nil(t, err) {
					bytesRead, err = r1.Read(p)
					assert.Equal(t, 0, bytesRead)
					if !assert.Equal(t, io.EOF, err) {
						r1.Close()
					}
				}
			}
		}()
		go func() {
			defer wg.Done()
			p := make([]byte, 7)
			bytesRead, err := r2.Read(p)
			assert.Equal(t, 6, bytesRead)
			assert.Equal(t, []byte("foobar"), p[0:bytesRead])
			if assert.Nil(t, err) {
				bytesRead, err = r2.Read(p)
				assert.Equal(t, 0, bytesRead)
				if !assert.Equal(t, io.EOF, err) {
					r2.Close()
				}
			}
		}()
		go func() {
			defer wg.Done()
			p := make([]byte, 4)
			bytesRead, err := r3.Read(p)
			assert.Equal(t, 4, bytesRead)
			assert.Equal(t, []byte("foob"), p[0:bytesRead])
			assert.Nil(t, err)
			r3.Close()
		}()
		wg.Wait()
	})
}

func Test_reader_Close(t *testing.T) {
	t.Run("Closing_twice", func(t *testing.T) {
		mr := NewMulteeReader(strings.NewReader("foo"))
		r := mr.NewReader()
		err := r.Close()
		if assert.NoError(t, err) {
			err = r.Close()
			assert.ErrorIs(t, err, ErrClosed)
		}
	})
}
