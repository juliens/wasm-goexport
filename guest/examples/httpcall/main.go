package main

import (
	"fmt"
	"net/http"

	"github.com/juliens/wasm-goexport/guest"
	_ "github.com/stealthrocket/net/http"
	"github.com/tetratelabs/wazero/api"
)

func main() {
	guest.SetExports([]*guest.Function{
		{
			ModuleName: "",
			Name:       "httpCall",
			Fn: func(i uint64) {
				resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", i))
				if err != nil {
					panic(err)
				}
				if resp.StatusCode != 200 {
					panic(fmt.Sprintf("http status code %d", resp.StatusCode))
				}
			},
			Params:  []api.ValueType{api.ValueTypeI64},
			Results: []api.ValueType{},
		},
	},
	)
}
