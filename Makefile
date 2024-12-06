BUILD_IMAGE?=k6-wrpc
BUILD_TAG?=latest

all: build

k6: *.go
	xk6 build --with xk6-wrpc=. --with github.com/grafana/xk6-dashboard@latest --with github.com/szkiba/xk6-top@latest --with github.com/cosmonic-labs/xk6-nats@latest

bindgen:
	wit-deps && wit-bindgen-wrpc go --out-dir internal --package $(shell go list)/internal wit
	(cd blaster-component && wit-deps && go generate)

component:
	(cd blaster-component && tinygo build -target wasip2 -wit-package wit -wit-world server)

build: k6 component

docker:
	docker build -t $(BUILD_IMAGE):$(BUILD_TAG) .

.PHONY: build bindgen docker
