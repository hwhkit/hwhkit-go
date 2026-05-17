// Package shutdown provides a context-backed token broadcast to background tasks.
package shutdown

import (
	"context"
	"sync"
)

type Token struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

func New() *Token {
	ctx, cancel := context.WithCancel(context.Background())
	return &Token{ctx: ctx, cancel: cancel}
}

func (t *Token) Context() context.Context { return t.ctx }

func (t *Token) Done() <-chan struct{} { return t.ctx.Done() }

func (t *Token) Cancel() {
	t.once.Do(t.cancel)
}

func (t *Token) Cancelled() bool {
	select {
	case <-t.ctx.Done():
		return true
	default:
		return false
	}
}
