# Multee, an io.Reader multiplexer

[![Go Reference](https://pkg.go.dev/badge/github.com/ComaVN/multee.svg)](https://pkg.go.dev/github.com/ComaVN/multee)
[![Keep a Changelog v1.1.0 badge][changelog-badge]][changelog]

## Purpose

This package implements a multiplexer for `io.Reader`, making it possible to read from a single `io.Reader` several times concurrently,
without needing to Seek back to the beginning.

## Usage

Create a multee-reader from a single `io.Reader`, and create as many readers as you need:
```go
	inputReader := strings.NewReader("Foo")
	mr := multee.NewMulteeReader(inputReader)
	r1 := mr.NewReader()
	r2 := mr.NewReader()
	r3 := mr.NewReader()
```

Now, you can use `r1`, `r2` and `r3` as a regular `io.ReadCloser`.

Each reader must be read in its own go-routine, and they must either be read until EOF or `Close()` must be called, or the MulteeReader will block.

The returned readers themselves are *not* concurrency-safe.

See also the [code examples][examples].

## Testing

Just run poor man's CI, `make test`.

## Contribute

Feel free to contribute, even if it's just to complain! Issues and pull requests are welcome.

See the [contributing instructions][contributing] for help to get started.


[changelog]: /CHANGELOG.md
[changelog-badge]: https://img.shields.io/badge/changelog-Keep%20a%20Changelog%20v1.1.0-%23E05735
[examples]: /examples
[contributing]: /CONTRIBUTING.md
