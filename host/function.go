package host

import (
	"context"

	"github.com/tetratelabs/wazero/api"
)

type function struct {
	api.Function
	errCh        chan error
	name         string
	params       []uint64
	results      []uint64
	callbackNum  uint32
	callbackChan chan uint32
	feedbackChan chan struct{}
	mod          *Module
	def          definition
}

func (l *function) Definition() api.FunctionDefinition {
	return l.def
}

func (l *function) Call(ctx context.Context, params ...uint64) ([]uint64, error) {
	l.params = params
	*l.mod.ptrCtx = context.WithValue(ctx, localFn_key, l)
	l.callbackChan <- l.callbackNum
	select {
	case <-l.feedbackChan:
		return l.results, nil
	case err := <-l.errCh:
		return nil, err
	}
}

func (l *function) CallWithStack(ctx context.Context, stack []uint64) error {
	// TODO implement me
	panic("implement me")
}

type GuestFunction struct {
	ModuleName string `json:"moduleName"`
	Name       string `json:"name"`
	Params     []api.ValueType
	Results    []api.ValueType
	Fn         func() `json:"-"`
}
