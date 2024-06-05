package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/juliens/wasm-goexport/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func main() {
	runtime := host.NewRuntime(wazero.NewRuntime(context.Background()))

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

	_, err = runtime.NewHostModuleBuilder("tata").NewFunctionBuilder().WithGoFunction(api.GoFunc(func(ctx context.Context, stack []uint64) {
		fmt.Println(ctx.Value("totototo"))
	}), []api.ValueType{}, []api.ValueType{}).Export("tata").Instantiate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	e, err := runtime.InstantiateModule(context.Background(), mod, wazero.NewModuleConfig())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), "totototo", "xxx")
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
