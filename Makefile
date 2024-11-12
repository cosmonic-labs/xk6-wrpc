BUILD_IMAGE?=k6-wrpc
BUILD_TAG?=latest

all: build

k6: *.go
	xk6 build --with xk6-wrpc=.

bindgen:
	wit-bindgen-wrpc go --out-dir internal --package $(shell go list)/internal wit

build: k6

run: k6
	./k6 run ./_examples/basic.js

docker:
	docker build -t $(BUILD_IMAGE):$(BUILD_TAG) .

.PHONY: build bindgen run docker
