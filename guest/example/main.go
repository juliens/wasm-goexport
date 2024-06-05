package main

import (
	"fmt"

	"github.com/juliens/wasm-goexport/guest"
	"github.com/tetratelabs/wazero/api"
)

func main() {
	guest.SetExports([]*guest.Function{
		{
			ModuleName: "try",
			Name:       "toto",
			Fn: func(i uint64) uint64 {
				fmt.Println("TOTO IS CALLED", i)
				return i * 2
			},
			Params:  []api.ValueType{api.ValueTypeI64},
			Results: []api.ValueType{api.ValueTypeI64},
		},
	},
	)
}
