# Multee, an io.Reader multiplexer

[![Go Reference](https://pkg.go.dev/badge/github.com/ComaVN/multee.svg)](https://pkg.go.dev/github.com/ComaVN/multee)
[![Keep a Changelog v1.1.0 badge][changelog-badge]][changelog]

## Purpose

This package implements a multiplexer for io.Readers, making it possible to read from a single io.Reader several times concurrently,
without needing to Seek back to the beginning.

## Usage

See the [code examples][examples]

## Testing

Just run poor man's CI, `make test`.

## Contribute

Feel free to contribute, even if it's just to complain! Issues and pull requests are welcome.

See the [contributing instructions][contributing] for help to get started.


[changelog]: /CHANGELOG.md
[changelog-badge]: https://img.shields.io/badge/changelog-Keep%20a%20Changelog%20v1.1.0-%23E05735
[examples]: /examples
[contributing]: /CONTRIBUTING.md
