package host

import (
	"context"
)

type magicContext struct {
	context.Context
}

func (m magicContext) Value(key any) any {
	val := GetRealCtx(m.Context).Value(key)
	if val != nil {
		return val
	}

	return m.Context.Value(key)
}
