all: build

k6:
	xk6 build --with xk6-wrpc=.

bindgen:
	wit-bindgen-wrpc go --out-dir internal --package $(shell go list)/internal wit

build: bindgen k6
	./k6

run: k6
	./k6 run ./_examples/basic.js

.PHONY: build bindgen run
