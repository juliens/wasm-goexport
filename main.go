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
	runtime := wazero.NewRuntime(context.Background())
	exporter := host.NewExporter(runtime)
	wasm, err := os.ReadFile("./main.wasm")
	if err != nil {
		log.Fatal(err)
	}
	instantiate, err := wasi_snapshot_preview1.Instantiate(context.Background(), runtime)
	if err != nil {
		log.Fatal(err)
	}
	_ = instantiate
	mod, err := runtime.CompileModule(context.Background(), wasm)
	if err != nil {
		log.Fatal(err)
	}
	e, err := exporter.InstantiateModule(context.Background(), mod, wazero.NewModuleConfig())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), "totototo", "fdsdfdsf")
	ret, err := e.ExportedFunction("toto").Call(ctx, 42)
	fmt.Println(ret)
	if err != nil {
		fmt.Println(err)
		return
	}
	ret, err = e.ExportedFunction("titi").Call(ctx, 45, 32)
	fmt.Println(ret)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(e.ExportedFunctionDefinitions())
}
