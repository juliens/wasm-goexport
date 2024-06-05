main.wasm: guest/example/main.go
	GOOS=wasip1 GOARCH=wasm go build -o ./main.wasm ./guest/example/main.go

.PHONY=build
build: main.wasm

test: build
	go test ./host