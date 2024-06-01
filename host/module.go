package host

import (
	"context"
	"encoding/json"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const HostModule = "go_exporter"

type Exporter struct {
	runtime wazero.Runtime
}

type definition struct {
	api.FunctionDefinition
	fn Function
}

func (d definition) ModuleName() string {
	return d.fn.ModuleName
}

func (d definition) Index() uint32 {
	// TODO implement me
	panic("implement me")
}

func (d definition) Import() (moduleName, name string, isImport bool) {
	// TODO implement me
	panic("implement me")
}

func (d definition) ExportNames() []string {
	// TODO implement me
	panic("implement me")
}

func (d definition) wazeroOnly() {
	// TODO implement me
	panic("implement me")
}

func (d definition) Name() string {
	return ""
}

func (d definition) DebugName() string {
	// TODO implement me
	panic("implement me")
}

func (d definition) GoFunction() interface{} {
	// TODO implement me
	panic("implement me")
}

func (d definition) ParamTypes() []api.ValueType {
	return d.fn.Params
}

func (d definition) ParamNames() []string {
	// TODO implement me
	panic("implement me")
}

func (d definition) ResultTypes() []api.ValueType {
	return d.fn.Results
}

func (d definition) ResultNames() []string {
	// TODO implement me
	panic("implement me")
}

type localFunc struct {
	api.Function
	errCh        chan error
	name         string
	params       []uint64
	results      []uint64
	callbackNum  uint32
	callbackChan chan uint32
	feedbackChan chan struct{}
	mod          *Module
	def          definition
}

func (l *localFunc) Definition() api.FunctionDefinition {
	return l.def
}

func (l *localFunc) Call(ctx context.Context, params ...uint64) ([]uint64, error) {
	l.params = params
	*l.mod.ptrCtx = context.WithValue(ctx, localFn, l)
	l.callbackChan <- l.callbackNum
	select {
	case <-l.feedbackChan:
		return l.results, nil
	case err := <-l.errCh:
		return nil, err
	}
}

func (l *localFunc) CallWithStack(ctx context.Context, stack []uint64) error {
	// TODO implement me
	panic("implement me")
}

func (l *localFunc) wazeroOnly() {
	// TODO implement me
	panic("implement me")
}

type Function struct {
	ModuleName string `json:"moduleName"`
	Name       string `json:"name"`
	Params     []api.ValueType
	Results    []api.ValueType
	Fn         func() `json:"-"`
}

type Module struct {
	api.Module
	exportedFn   map[string]*localFunc
	callbackChan chan uint32
	ptrCtx       *context.Context
}

func (m *Module) ExportedFunction(name string) api.Function {
	if fn := m.Module.ExportedFunction(name); fn != nil {
		return fn
	}
	if f, ok := m.exportedFn[name]; ok {
		return f
	}
	return nil
}

func (m *Module) ExportedFunctionDefinitions() map[string]api.FunctionDefinition {
	exportedFns := make(map[string]api.FunctionDefinition)
	for key, functionDefinition := range m.Module.ExportedFunctionDefinitions() {
		exportedFns[key] = functionDefinition
	}

	for key, functionDefinition := range m.exportedFn {
		exportedFns[key] = functionDefinition.Definition()
	}
	return exportedFns
}

func NewExporter(runtime wazero.Runtime) *Exporter {
	return &Exporter{
		runtime: runtime,
	}
}

func (e *Exporter) detectGoExports(mod wazero.CompiledModule) bool {
	for _, f := range mod.ImportedFunctions() {
		if module, _, _ := f.Import(); module == HostModule {
			return true
		}
	}
	return false
}

func (e Exporter) InstantiateModule(ctx context.Context, compiled wazero.CompiledModule, config wazero.ModuleConfig) (api.Module, error) {
	if !e.detectGoExports(compiled) {
		return e.runtime.InstantiateModule(ctx, compiled, config)
	}

	exportsChan := make(chan []Function)
	callbackChan := make(chan uint32)
	feedbackChan := make(chan struct{})
	b := e.runtime.NewHostModuleBuilder(HostModule).
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			buf := uint32(stack[0])
			bufLen := uint32(stack[1])
			data, _ := mod.Memory().Read(buf, bufLen)
			exportedFn := []Function{}
			err := json.Unmarshal(data, &exportedFn)
			if err != nil {
				panic(err)
			}
			exportsChan <- exportedFn
		}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("set_exports").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			stack[0] = uint64(<-callbackChan)
		}), []api.ValueType{}, []api.ValueType{api.ValueTypeI32}).Export("get_callback").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			fn := getLocalFunc(ctx)
			index := stack[0]
			stack[0] = fn.params[index]
		}), []api.ValueType{api.ValueTypeI64}, []api.ValueType{api.ValueTypeI64}).Export("get_arg").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			fn := getLocalFunc(ctx)
			if len(stack) > 0 {
				fn.results = []uint64{uint64(stack[0])}
			}

		}), []api.ValueType{api.ValueTypeI64}, []api.ValueType{}).Export("set_result").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			feedbackChan <- struct{}{}
		}), []api.ValueType{}, []api.ValueType{}).Export("wait_feedback")
	b.Instantiate(context.Background())
	var mod api.Module
	errCh := make(chan error)

	// To create a pointer for context.Context
	cpCtx := context.WithValue(ctx, ctx_cp, struct{}{})
	var ptrCtx = &cpCtx
	ctx = context.WithValue(ctx, ctx_key, ptrCtx)

	go func() {
		var err error
		mod, err = e.runtime.InstantiateModule(ctx, compiled, config.WithStartFunctions().WithStdout(os.Stdout))
		if err != nil {
			errCh <- err
			return
		}

		_, err = mod.ExportedFunction("_start").Call(ctx)

		if err != nil {
			errCh <- err
		}
	}()

	modu := &Module{callbackChan: callbackChan, ptrCtx: ptrCtx}
	select {
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case exportedFn := <-exportsChan:
		exported := map[string]*localFunc{}
		for i, f := range exportedFn {
			exported[f.Name] = &localFunc{
				errCh:        errCh,
				callbackNum:  uint32(i),
				callbackChan: callbackChan,
				feedbackChan: feedbackChan,
				mod:          modu,
				def:          definition{fn: f},
			}
		}
		modu.exportedFn = exported
		modu.Module = mod
		return modu, nil
	}
}

type key_type string

var ctx_key = key_type("key")
var ctx_cp = key_type("cp")
var localFn = key_type("fn")

func GetRealCtx(ctx context.Context) context.Context {
	curCtx := ctx.Value(ctx_key)
	if realCtx, ok := curCtx.(*context.Context); ok {
		return *realCtx
	}
	return ctx

}

func getLocalFunc(ctx context.Context) *localFunc {
	ctx = GetRealCtx(ctx)
	if fn, ok := ctx.Value(localFn).(*localFunc); ok {
		return fn
	}
	return nil
}
