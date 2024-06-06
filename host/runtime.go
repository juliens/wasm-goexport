package host

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type key_type string

var ctx_key = key_type("key")
var ctx_cp_key = key_type("cp")
var localFn_key = key_type("fn")
var callback_key = key_type("callback")

func NewRuntime(r wazero.Runtime) wazero.Runtime {
	return Runtime{Runtime: r, exportsChan: make(chan []GuestFunction)}
}

type Runtime struct {
	wazero.Runtime
	exportsChan chan []GuestFunction
}

func (e Runtime) InstantiateModule(ctx context.Context, compiled wazero.CompiledModule, config wazero.ModuleConfig) (api.Module, error) {
	if !DetectGoExports(compiled) {
		return e.Runtime.InstantiateModule(ctx, compiled, config)
	}

	err := e.buildHost(ctx)
	if err != nil {
		return nil, err
	}

	mod, err := e.Runtime.InstantiateModule(ctx, compiled, config.WithStartFunctions())
	if err != nil {
		return nil, err
	}

	if mod.ExportedFunction("_start") == nil {
		return mod, nil
	}

	callbackChan := make(chan uint32)
	feedbackChan := make(chan struct{})
	errCh := make(chan error)

	// To create a pointer for context.Context
	cpCtx := context.WithValue(ctx, ctx_cp_key, struct{}{})
	var ptrCtx = &cpCtx
	ctx = context.WithValue(ctx, ctx_key, ptrCtx)
	ctx = context.WithValue(ctx, callback_key, callbackChan)

	go func() {
		ctx = magicContext{ctx}
		_, err = mod.ExportedFunction("_start").Call(ctx)

		if err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case exportedFn := <-e.exportsChan:
		modu := &Module{callbackChan: callbackChan, ptrCtx: ptrCtx, Module: mod}

		exported := map[string]*function{}
		for i, f := range exportedFn {
			exported[f.Name] = &function{
				errCh:        errCh,
				callbackNum:  uint32(i),
				callbackChan: callbackChan,
				feedbackChan: feedbackChan,
				mod:          modu,
				def:          definition{fn: f},
			}

		}
		modu.exportedFn = exported
		return modu, nil
	}
}
