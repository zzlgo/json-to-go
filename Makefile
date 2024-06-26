TINYGOROOT := $(shell tinygo env TINYGOROOT)

.PHONY: build test linter

build:
	rm -rf static/json-to-go/wasm_exec.js
	rm -rf static/json-to-go/main.wasm
	cp $(TINYGOROOT)/targets/wasm_exec.js static/json-to-go
    # error: could not find wasm-opt, set the WASMOPT environment variable to override。（brew install binaryen fixed it）
	# -no-debug 减少大小  -stack-size=1MB递归调用栈不够
	tinygo build -gc=leaking -no-debug -stack-size=1MB -panic=trap -o static/json-to-go/main-pre.wasm -target wasm cmd/wasm/main.go
	wasm-opt -Os static/json-to-go/main-pre.wasm -o static/json-to-go/main.wasm
	rm -rf static/json-to-go/main-pre.wasm

docker: build
	docker build -t json-to-go .
test:
	go test -v -cover -race

linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
	golangci-lint cache clean && go clean -modcache -cache -i
	golangci-lint --version
	GOARCH=wasm GOOS=js golangci-lint run --timeout=10m
