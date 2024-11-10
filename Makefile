.PHONY: clean
clean:
	git clean -xdff

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	test/test.sh

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: update
update:
	go get -u
