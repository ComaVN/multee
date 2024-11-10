# Contributing

## Setup your development environment

### Requirements

We use [asdf](https://asdf-vm.com/) for installing [required tools](.tool-versions):
```sh
asdf plugin add golang
asdf plugin add golangci-lint
asdf plugin add pre-commit
asdf install
```

We use pre-commmit for providing a git hook:
```sh
pre-commit install
```

### Testing

To run all tests:
```sh
make test
```
