package guest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/tetratelabs/wazero/api"
)

type Function struct {
	ModuleName string `json:"moduleName"`
	Name       string `json:"name"`
	Params     []api.ValueType
	Results    []api.ValueType
	Fn         any `json:"-"`
	fn         reflect.Value
}

func run(exports []*Function) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	for {
		callback := getCallback()
		if len(exports) < int(callback) {
			return
		}
		fn := exports[callback]
		params := []reflect.Value{}
		for i := range fn.Params {
			params = append(params, reflect.ValueOf(get_arg(uint64(i))))
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
		panic("TEST")
	}
	setExports(b)
	for _, export := range exports {
		if reflect.TypeOf(export.Fn).Kind() != reflect.Func {
			fmt.Println("PANIC")
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
