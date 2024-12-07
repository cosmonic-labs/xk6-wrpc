BUILD_IMAGE?=k6-wrpc
BUILD_TAG?=latest

all: build

k6:
	xk6 build --with xk6-wrpc=. --with github.com/grafana/xk6-dashboard@latest --with github.com/szkiba/xk6-top@latest --with github.com/cosmonic-labs/xk6-nats@latest

bindgen: k6-bindgen components-bindgen

k6-bindgen:
	wit-deps && wit-bindgen-wrpc go --out-dir internal --package $(shell go list)/internal wit
components-bindgen:
	@for component in components/*; do\
		echo "==> $${component}";\
		make -C $${component} bindgen;\
	done

components:
	@for component in components/*; do\
		echo "==> $${component}";\
		make -C $${component} build;\
	done

build: k6 components

docker:
	docker build -t $(BUILD_IMAGE):$(BUILD_TAG) .

.PHONY: build k6 bindgen docker components components-bindgen
