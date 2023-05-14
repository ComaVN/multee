// Copyright 2023 Roel Harbers.
// Use of this source code is governed by the BEER-WARE license
// that can be found in the LICENSE file.

package multee_test

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"testing"

	"github.com/ComaVN/multee"
)

func Test_multee_monkeytest(t *testing.T) {
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
			mr := multee.NewMulteeReader(inputR)
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
