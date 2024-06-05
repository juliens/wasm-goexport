package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/juliens/wasm-goexport/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func main() {
	runtime := host.NewRuntime(wazero.NewRuntime(context.Background()))

	wasm, err := os.ReadFile("./main.wasm")
	if err != nil {
		log.Fatal(err)
	}

	_, err = wasi_snapshot_preview1.Instantiate(context.Background(), runtime)
	if err != nil {
		log.Fatal(err)
	}

	mod, err := runtime.CompileModule(context.Background(), wasm)
	if err != nil {
		log.Fatal(err)
	}

	e, err := runtime.InstantiateModule(context.Background(), mod, wazero.NewModuleConfig())
	if err != nil {
		log.Fatal(err)
	}

	ret, err := e.ExportedFunction("toto").Call(context.Background(), 42)
	fmt.Println(ret)
	if err != nil {
		fmt.Println(err)
		return
	}
}
