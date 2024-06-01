package guest

import (
	"reflect"

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
