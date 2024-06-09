package host

import (
	"context"
	"encoding/json"

	"github.com/tetratelabs/wazero/api"
)

const HostModule = "go_exporter"

func (e Runtime) buildHost(ctx context.Context) error {
	if e.Runtime.Module(HostModule) != nil {
		return nil
	}
	_, err := e.Runtime.NewHostModuleBuilder(HostModule).
		NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			buf := uint32(stack[0])
			bufLen := uint32(stack[1])
			data, _ := mod.Memory().Read(buf, bufLen)
			exportedFn := []GuestFunction{}
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
