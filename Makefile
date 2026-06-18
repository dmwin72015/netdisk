APP_VERSION ?= $(shell git describe --tags --always 2>/dev/null)

build:
	APP_VERSION=$(APP_VERSION) docker compose build

.PHONY: build
