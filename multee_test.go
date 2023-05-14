// Copyright 2023 Roel Harbers.
// Use of this source code is governed by the BEER-WARE license
// that can be found in the LICENSE file.

package multee

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
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
		mr.read(make([]byte, 3), 1)
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

func Test_reader_monkeytest(t *testing.T) {
	for rndSeed := int64(0); rndSeed < 10; rndSeed++ { // Run the test with 10 different, predictable seeds.
		t.Run(fmt.Sprintf("Monkey_test_with_rnd_seed_%d", rndSeed), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(rndSeed))
			inputR, inputW := io.Pipe()
			go func(rndSeed int64) {
				// Generate an infinite stream of bytes, each 8 bytes containing their 64-bit offset in the stream.
				buf := make([]byte, 8)
				offs := uint64(0)
				for {
					binary.LittleEndian.PutUint64(buf, offs)
					l, err := inputW.Write(buf)
					if l != len(buf) || err != nil {
						panic(fmt.Errorf("rnd seed %d: failed to write the full uint64 (%d, only %d bytes written, err: %v)", rndSeed, offs, l, err))
					}
					offs += uint64(l)
				}
			}(rndSeed)
			mr := NewMulteeReader(inputR)
			readers := make([]io.ReadCloser, rnd.Intn(20)) // Use between 0 and 19 (inclusive) readers.
			for idx := range readers {
				readers[idx] = mr.NewReader()
			}
			var wg sync.WaitGroup
			wg.Add(len(readers))
			for rdrIdx, r := range readers {
				numReads := rnd.Intn(20)         // Do between 0 and 19 (inclusive) reads.
				readLen := 1 + rnd.Intn(64*1024) // Read between 1B and 64KiB (inclusive) at a time.
				go func(r io.ReadCloser, rndSeed int64, rdrIdx int) {
					// This reads a random number of randomly sized chunks
					// and checks if the read bytes match their offset, as the writer wrote them.
					defer wg.Done()
					defer r.Close()
					var prevBuf []byte // This keeps the remainder of bytes, that aren't in a multiple of 8.
					offs := uint64(0)
					for i := 0; i < numReads; i++ {
						buf := make([]byte, len(prevBuf)+readLen)
						copy(buf, prevBuf)
						l, err := io.ReadFull(r, buf[len(prevBuf):])
						if l != len(buf)-len(prevBuf) || err != nil {
							t.Errorf("rnd seed %d, reader index: %d: failed to read the full buffer (len %d, %d bytes read, err: %v)", rndSeed, rdrIdx, len(buf), l, err)
						}
						bufOffs := 0
						for ; bufOffs <= len(buf)-8; bufOffs += 8 {
							got := binary.LittleEndian.Uint64(buf[bufOffs : bufOffs+8])
							if got != offs {
								t.Errorf("rnd seed %d, reader index: %d: expected to read offset %d, got %d", rndSeed, rdrIdx, offs, got)
								return
							}
							offs += 8
						}
						prevBuf = buf[bufOffs:]
					}
				}(r, rndSeed, rdrIdx)
			}
			wg.Wait()
		})
	}
}
