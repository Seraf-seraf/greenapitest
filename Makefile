.PHONY: run wasm build

run: wasm
	go run ./cmd/server

wasm:
	GOOS=js GOARCH=wasm go build -o web/assets/main.wasm ./cmd/wasm

build: wasm
	go build ./...
