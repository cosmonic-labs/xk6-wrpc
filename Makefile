all: build

k6: *.go
	xk6 build --with xk6-wrpc=.

bindgen:
	wit-bindgen-wrpc go --out-dir internal --package $(shell go list)/internal wit

build: k6

run: k6
	./k6 run ./_examples/basic.js

.PHONY: build bindgen run
