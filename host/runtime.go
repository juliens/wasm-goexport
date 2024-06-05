package host

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

func NewRuntime(r wazero.Runtime) wazero.Runtime {
	runtime := Runtime{Runtime: r, exportsChan: make(chan []Function)}
	// TODO conditionnal ?
	runtime.buildHost(context.Background())
	return runtime
}

type Runtime struct {
	wazero.Runtime
	exportsChan chan []Function
}

func (e *Runtime) buildHost(ctx context.Context) error {
	_, err := e.Runtime.NewHostModuleBuilder(HostModule).
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
			e.exportsChan <- exportedFn
		}), []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{}).Export("set_exports").
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			callback := ctx.Value(callback_key).(chan uint32)
			stack[0] = uint64(<-callback)
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
			fn := getLocalFunc(ctx)
			fn.feedbackChan <- struct{}{}
		}), []api.ValueType{}, []api.ValueType{}).Export("wait_feedback").Instantiate(ctx)
	return err
}

type proxyCompiledModule struct {
	wazero.CompiledModule
	fnDef map[string]api.FunctionDefinition
}

func (p proxyCompiledModule) ExportedFunctionDefinitions() map[string]api.FunctionDefinition {
	return p.fnDef
}

func (e Runtime) CompileModule(ctx context.Context, binary []byte) (wazero.CompiledModule, error) {
	mod, err := e.Runtime.CompileModule(ctx, binary)
	if err != nil {
		return nil, err
	}
	if !DetectGoExports(mod) {
		return mod, nil
	}

	iMod, err := e.InstantiateModule(ctx, mod, wazero.NewModuleConfig().WithStartFunctions())
	if err != nil {
		return nil, err
	}
	def := iMod.ExportedFunctionDefinitions()
	iMod.Close(ctx)
	return proxyCompiledModule{CompiledModule: mod, fnDef: def}, nil
}

func (e Runtime) InstantiateModule(ctx context.Context, compiled wazero.CompiledModule, config wazero.ModuleConfig) (api.Module, error) {
	callbackChan := make(chan uint32)
	feedbackChan := make(chan struct{})

	var mod api.Module
	errCh := make(chan error)

	// To create a pointer for context.Context
	cpCtx := context.WithValue(ctx, ctx_cp_kety, struct{}{})
	var ptrCtx = &cpCtx
	ctx = context.WithValue(ctx, ctx_key, ptrCtx)
	ctx = context.WithValue(ctx, callback_key, callbackChan)
	go func() {
		var err error
		ctx := magicContext{ctx}
		mod, err = e.Runtime.InstantiateModule(ctx, compiled, config.WithStartFunctions())
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
	case exportedFn := <-e.exportsChan:
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

type magicContext struct {
	ctx context.Context
}

func (m magicContext) Deadline() (deadline time.Time, ok bool) {
	return m.ctx.Deadline()
}

func (m magicContext) Done() <-chan struct{} {
	return m.ctx.Done()
}

func (m magicContext) Err() error {
	return m.ctx.Err()
}

func (m magicContext) Value(key any) any {
	val := m.ctx.Value(key)
	if val != nil {
		return val
	}
	return GetRealCtx(m.ctx).Value(key)
}
