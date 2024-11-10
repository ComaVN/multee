// Copyright 2023-2024 Roel Harbers.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Example code for how to use an alternative implementation of the multee package.
package main

import (
	"fmt"
	"io"
	"strings"
	"sync"

	multee "github.com/ComaVN/multee/alt/byteslicechan"
)

func main() {
	// When you have this stream of input data...
	inputReader := strings.NewReader("Foo")
	// ...and these are things you want to do with that data...
	f1 := func(r io.ReadCloser) {
		// Print in UPPERCASE
		b, _ := io.ReadAll(r)
		fmt.Println(strings.ToUpper(string(b)))
	}
	f2 := func(r io.ReadCloser) {
		// Print in LOWERCASE
		b, _ := io.ReadAll(r)
		fmt.Println(strings.ToLower(string(b)))
	}
	f3 := func(r io.ReadCloser) {
		// Print as-is
		b, _ := io.ReadAll(r)
		fmt.Println(string(b))
	}
	// ...you use Multee...
	mr := multee.NewMulteeReader(inputReader)
	r1 := mr.NewReader()
	r2 := mr.NewReader()
	r3 := mr.NewReader()
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
