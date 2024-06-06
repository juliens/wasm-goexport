SOURCES := $(wildcard guest/examples/*/*.go)

.PHONY=build
build: $(SOURCES)
	@for f in $^; do \
	    GOOS=wasip1 GOARCH=wasm go build -o $$(echo $$f | sed -e 's/\.go/\.wasm/') $$f; \
	done
test: build
	go test ./host