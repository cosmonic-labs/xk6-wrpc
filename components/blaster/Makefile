all: bindgen build

.PHONY: bindgen
bindgen:
	wit-deps
	go generate

.PHONY: build
build:
	tinygo build -target wasip2 -wit-package wit -wit-world server -o blaster.wasm
