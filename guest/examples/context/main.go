package main

import (
	"github.com/juliens/wasm-goexport/guest"
	"github.com/tetratelabs/wazero/api"
)

func main() {
	guest.SetExports([]*guest.Function{
		{
			ModuleName: "",
			Name:       "callImport",
			Fn: func() {
				verifyContext()
			},
			Params:  []api.ValueType{},
			Results: []api.ValueType{},
		},
	},
	)
}

//go:wasmimport main verifyContext
func verifyContext()
