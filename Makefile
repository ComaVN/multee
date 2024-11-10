.PHONY: clean
clean:
	git clean -xdff

.PHONY: lint
lint:
	golangci-lint run

# TODO: Poor man's automated testing, replace this with some proper CI
.PHONY: test
test: unit-test integration-test compat-test
	@echo "All tests Ok"

.PHONY: unit-test
unit-test:
	test/unit-test.sh

.PHONY: integration-test
integration-test:
	test/example-test.sh

.PHONY: compat-test
compat-test:
	test/compat-test.sh

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: update
update:
	go get -u
