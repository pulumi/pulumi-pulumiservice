# Run the linters in .golangci.local.yml as part of `make lint` and
# `make lint.fix`. They live in their own config because ci-mgmt regenerates
# .golangci.yml and the Makefile.
lint: lint.local
lint.fix: lint.local.fix

lint.local:
	cd provider && golangci-lint run --path-prefix provider -c ../.golangci.local.yml

lint.local.fix:
	cd provider && golangci-lint run --path-prefix provider -c ../.golangci.local.yml --fix

.PHONY: lint.local lint.local.fix
