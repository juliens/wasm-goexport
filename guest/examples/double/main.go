package main

import (
	"github.com/juliens/wasm-goexport/guest"
	"github.com/tetratelabs/wazero/api"
)

func main() {
	guest.SetExports([]*guest.Function{
		{
			ModuleName: "",
			Name:       "double",
			Fn: func(i uint64) uint64 {
				return i * 2
			},
			Params:  []api.ValueType{api.ValueTypeI64},
			Results: []api.ValueType{api.ValueTypeI64},
		},
	},
	)
}
