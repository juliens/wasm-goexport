package host

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func TestE2EDouble(t *testing.T) {
	runtime := NewRuntime(wazero.NewRuntime(context.Background()))

	wasm, err := os.ReadFile("../main.wasm")
	require.NoError(t, err)

	_, err = wasi_snapshot_preview1.Instantiate(context.Background(), runtime)
	require.NoError(t, err)

	mod, err := runtime.CompileModule(context.Background(), wasm)
	require.NoError(t, err)

	_, err = runtime.NewHostModuleBuilder("main").NewFunctionBuilder().WithGoFunction(api.GoFunc(func(ctx context.Context, stack []uint64) {
		val := ctx.Value(test_key{})
		require.NotNil(t, val)
	}), []api.ValueType{}, []api.ValueType{}).Export("verifyContext").Instantiate(context.Background())
	require.NoError(t, err)

	e, err := runtime.InstantiateModule(context.Background(), mod, wazero.NewModuleConfig())
	require.NoError(t, err)

	ret, err := e.ExportedFunction("double").Call(context.Background(), 42)
	require.Len(t, ret, 1)
	require.EqualValues(t, 84, ret[0])
	require.NoError(t, err)

}

func TestE2EContext(t *testing.T) {
	runtime := NewRuntime(wazero.NewRuntime(context.Background()))

	wasm, err := os.ReadFile("../main.wasm")
	require.NoError(t, err)

	_, err = wasi_snapshot_preview1.Instantiate(context.Background(), runtime)
	require.NoError(t, err)

	mod, err := runtime.CompileModule(context.Background(), wasm)
	require.NoError(t, err)

	call := false
	_, err = runtime.NewHostModuleBuilder("main").NewFunctionBuilder().WithGoFunction(api.GoFunc(func(ctx context.Context, stack []uint64) {
		val := ctx.Value(test_key{})
		require.NotNil(t, val)
		call = true
	}), []api.ValueType{}, []api.ValueType{}).Export("verifyContext").Instantiate(context.Background())
	require.NoError(t, err)

	e, err := runtime.InstantiateModule(context.Background(), mod, wazero.NewModuleConfig())
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), test_key{}, struct{}{})
	_, err = e.ExportedFunction("callImport").Call(ctx, 42)
	require.NoError(t, err)
	require.True(t, call)

}

type test_key struct{}
