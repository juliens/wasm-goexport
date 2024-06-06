package host

import "github.com/tetratelabs/wazero/api"

type definition struct {
	api.FunctionDefinition
	fn GuestFunction
}

func (d definition) ModuleName() string {
	return d.fn.ModuleName
}

func (d definition) Index() uint32 {
	// TODO implement me
	panic("implement me")
}

func (d definition) Import() (moduleName, name string, isImport bool) {
	// TODO implement me
	panic("implement me")
}

func (d definition) ExportNames() []string {
	// TODO implement me
	panic("implement me")
}

func (d definition) Name() string {
	return ""
}

func (d definition) DebugName() string {
	// TODO implement me
	panic("implement me")
}

func (d definition) GoFunction() interface{} {
	// TODO implement me
	panic("implement me")
}

func (d definition) ParamTypes() []api.ValueType {
	return d.fn.Params
}

func (d definition) ParamNames() []string {
	// TODO implement me
	panic("implement me")
}

func (d definition) ResultTypes() []api.ValueType {
	return d.fn.Results
}

func (d definition) ResultNames() []string {
	// TODO implement me
	panic("implement me")
}
