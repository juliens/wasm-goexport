//go:build wasip1

package guest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/tetratelabs/wazero/api"
)

func run(exports []*Function) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in run", r)
		}
	}()
	for {
		callback := getCallback()
		if len(exports) < int(callback) {
			return
		}
		fn := exports[callback]
		params := []reflect.Value{}

		for i, param := range fn.Params {
			var value reflect.Value
			switch param {
			case api.ValueTypeI64:
				value = reflect.ValueOf(get_arg(uint64(i)))
			case api.ValueTypeI32:
				value = reflect.ValueOf(uint32(get_arg(uint64(i))))
			}

			params = append(params, value)
		}
		callResults := fn.fn.Call(params)
		if len(callResults) > 0 {
			set_result(callResults[0].Uint())
		}
		wait_feedback()
	}
}

func getCallback() uint32 {
	return get_callback()
}

func SetExports(exports []*Function) {
	b, err := json.Marshal(exports)
	if err != nil {
		panic(err)
	}
	setExports(b)
	for _, export := range exports {
		if reflect.TypeOf(export.Fn).Kind() != reflect.Func {
			panic("not a function")
		}
		export.fn = reflect.ValueOf(export.Fn)

	}
	run(exports)
}

func setExports(b []byte) {
	ptr := unsafe.Pointer(unsafe.SliceData(b))
	set_exports(uint32(uintptr(ptr)), uint32(len(b)))
}

//go:wasmimport go_exporter set_exports
func set_exports(bufPtr uint32, bufLen uint32)

//go:wasmimport go_exporter get_callback
func get_callback() uint32

//go:wasmimport go_exporter wait_feedback
func wait_feedback()

//go:wasmimport go_exporter get_arg
func get_arg(i uint64) uint64

//go:wasmimport go_exporter set_result
func set_result(i uint64)
