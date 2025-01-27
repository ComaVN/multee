// Copyright 2023-2025 Roel Harbers.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Example code for how to use the multee package on huge input streams
// (ie. when it doesn't fit in memory).
package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"github.com/ComaVN/multee"
)

func main() {
	// When you have this infinite stream of input data...
	inputReader := rand.New(rand.NewSource(31337))
	// ...and these are things you want to do with a huge prefix of that data...
	f1 := func(r io.ReadCloser) {
		defer r.Close()
		// Output the average byte value of the first gigabyte data,
		p := make([]byte, 1024)
		var cnt uint64
		var sum uint64
		for cnt < 1024*1024*1024 {
			l, err := r.Read(p)
			cnt += uint64(l)
			for _, v := range p[:l] {
				sum += uint64(v)
			}
			if err != nil {
				fmt.Printf("Error reading for f1: %v\n", err)
				break
			}
		}
		if cnt <= 0 {
			fmt.Println("No bytes to be read for averaging")
			return
		}
		fmt.Printf("Average of first %d bytes: %d\n", cnt, sum/cnt)
	}
	f2 := func(r io.ReadCloser) {
		defer r.Close()
		// Output all occurences of 0xDEADBEEF in the first 32 gigabytes of data.
		p := make([]byte, 1024+3)
		var cnt uint64
		var found int
		for cnt < 32*1024*1024*1024 {
			l, err := r.Read(p[3:])
			cnt += uint64(l)
			idx := bytes.Index(p[:l+3], []byte{0xDE, 0xAD, 0xBE, 0xEF})
			if idx >= 0 {
				found++
				fmt.Printf("0xDEADBEEF found at index %d\n", cnt-uint64(l)-3+uint64(idx))
			}
			if err != nil {
				fmt.Printf("Error reading for f1: %v\n", err)
				break
			}
			copy(p[:3], p[len(p)-3:])
		}
		if found == 0 {
			fmt.Printf("0xDEADBEEF not found in first %d bytes\n", cnt)
		} else {
			fmt.Printf("0xDEADBEEF found %d times in first %d bytes\n", found, cnt)
		}
	}
	f3 := func(r io.ReadCloser) {
		defer r.Close()
		// Output the number of collections of at least 5 identical consecutive bytes in the first 16 gigabytes of data.
		var cnt uint64
		p := make([]byte, 1)
		l, err := r.Read(p)
		if err != nil || l != 1 {
			fmt.Printf("Error reading first byte for f3: %d, %v\n", l, err)
			return
		}
		cnt += uint64(l)
		collByte := p[0]
		collSize := 1
		var found int
		p = make([]byte, 1024)
		for cnt < 16*1024*1024*1024 {
			l, err := r.Read(p)
			cnt += uint64(l)
			for idx, v := range p[:l] {
				if v == collByte {
					collSize++
					if collSize == 5 {
						found++
						fmt.Printf("Collection of at least 5 consecutive %#x bytes found at index %d\n", collByte, cnt-uint64(l)+uint64(idx))
					}
				} else {
					collByte = v
					collSize = 1
				}
			}
			if err != nil {
				fmt.Printf("Error reading for f3: %v\n", err)
				break
			}
		}
		if found == 0 {
			fmt.Printf("No collections of at least 5 identical consecutive bytes found in first %d bytes\n", cnt)
		} else {
			fmt.Printf("%d collections of at least 5 identical consecutive bytes found in first %d bytes\n", found, cnt)
		}
	}
	// ...you use Multee...
	readers := multee.NewMulteeReader(inputReader)
	r1 := readers.NewReader()
	r2 := readers.NewReader()
	r3 := readers.NewReader()
	// ...and do those things!
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		f1(r1)
	}()
	go func() {
		defer wg.Done()
		f2(r2)
	}()
	go func() {
		defer wg.Done()
		f3(r3)
	}()
	// ...and wait for them to finish
	wg.Wait()
}
