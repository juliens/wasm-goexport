package guest

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/juliens/wasm-goexport/host"
	importswazergo "github.com/stealthrocket/wasi-go/imports"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed examples/double/main.wasm
var doubleSrc []byte

//go:embed examples/context/main.wasm
var contextSrc []byte

//go:embed examples/httpcall/main.wasm
var httpcallSrc []byte

func TestE2EDouble(t *testing.T) {
	runtime := host.NewRuntime(wazero.NewRuntime(context.Background()))

	_, err := wasi_snapshot_preview1.Instantiate(context.Background(), runtime)
	require.NoError(t, err)

	mod, err := runtime.CompileModule(context.Background(), doubleSrc)
	require.NoError(t, err)

	e, err := runtime.InstantiateModule(context.Background(), mod, wazero.NewModuleConfig())
	require.NoError(t, err)

	ret, err := e.ExportedFunction("double").Call(context.Background(), 42)
	require.Len(t, ret, 1)
	require.EqualValues(t, 84, ret[0])
	require.NoError(t, err)

}

func TestE2EContext(t *testing.T) {
	runtime := host.NewRuntime(wazero.NewRuntime(context.Background()))

	_, err := wasi_snapshot_preview1.Instantiate(context.Background(), runtime)
	require.NoError(t, err)

	mod, err := runtime.CompileModule(context.Background(), contextSrc)
	require.NoError(t, err)

	type test_key struct{}

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

func TestE2EHTTPCall(t *testing.T) {
	ctx := context.Background()
	runtime := host.NewRuntime(wazero.NewRuntime(ctx))

	mod, err := runtime.CompileModule(ctx, httpcallSrc)
	require.NoError(t, err)

	builder := importswazergo.NewBuilder().WithSocketsExtension("auto", mod)
	ctx, _, err = builder.Instantiate(ctx, runtime)
	require.NoError(t, err)

	e, err := runtime.InstantiateModule(ctx, mod, wazero.NewModuleConfig().WithStdout(os.Stdout))
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	parse, err := url.Parse(srv.URL)
	require.NoError(t, err)
	port, err := strconv.Atoi(parse.Port())
	require.NoError(t, err)

	_, err = e.ExportedFunction("httpCall").Call(ctx, uint64(port))
	require.NoError(t, err)
}
