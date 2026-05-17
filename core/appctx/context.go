// Package appctx is the type-keyed value store handed to integration providers and request handlers.
package appctx

import (
	"reflect"
	"sync"
)

type Context struct {
	mu    sync.RWMutex
	typed map[reflect.Type]any
	named map[string]any
}

func New() *Context {
	return &Context{
		typed: make(map[reflect.Type]any),
		named: make(map[string]any),
	}
}

func Insert[T any](c *Context, v *T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.typed[typeKey[T]()] = v
}

func Get[T any](c *Context) (*T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.typed[typeKey[T]()]
	if !ok {
		return nil, false
	}
	out, ok := v.(*T)
	return out, ok
}

func MustGet[T any](c *Context) *T {
	v, ok := Get[T](c)
	if !ok {
		panic("appctx: missing required value of type " + typeKey[T]().String())
	}
	return v
}

func (c *Context) InsertNamed(key string, v any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.named[key] = v
}

func (c *Context) GetNamed(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.named[key]
	return v, ok
}

func (c *Context) NamedKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, 0, len(c.named))
	for k := range c.named {
		out = append(out, k)
	}
	return out
}

func typeKey[T any]() reflect.Type {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		// Handles interface-typed T where zero is nil.
		return reflect.TypeOf((*T)(nil)).Elem()
	}
	return t
}
