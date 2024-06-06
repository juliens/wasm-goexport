package host

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type Module struct {
	api.Module
	exportedFn   map[string]*function
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

func DetectGoExports(mod wazero.CompiledModule) bool {
	for _, f := range mod.ImportedFunctions() {
		if module, _, _ := f.Import(); module == HostModule {
			return true
		}
	}
	return false
}

func GetRealCtx(ctx context.Context) context.Context {
	curCtx := ctx.Value(ctx_key)
	if realCtx, ok := curCtx.(*context.Context); ok {
		return *realCtx
	}
	return ctx

}

func getLocalFunc(ctx context.Context) *function {
	ctx = GetRealCtx(ctx)
	if fn, ok := ctx.Value(localFn_key).(*function); ok {
		return fn
	}
	return nil
}
