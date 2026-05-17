// Package runtime exposes the set of integration keys the running binary supports.
package runtime

import "sort"

type Features struct {
	enabled map[string]struct{}
}

func New() *Features {
	return &Features{enabled: make(map[string]struct{})}
}

func (f *Features) Enable(key string) *Features {
	if f.enabled == nil {
		f.enabled = make(map[string]struct{})
	}
	f.enabled[key] = struct{}{}
	return f
}

func (f *Features) EnableIf(key string, on bool) *Features {
	if on {
		f.Enable(key)
	}
	return f
}

func (f *Features) Contains(key string) bool {
	if f == nil || f.enabled == nil {
		return false
	}
	_, ok := f.enabled[key]
	return ok
}

func (f *Features) Keys() []string {
	if f == nil {
		return nil
	}
	out := make([]string, 0, len(f.enabled))
	for k := range f.enabled {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func KnownFeatures() []string {
	return []string{"postgres", "redis", "mongodb", "nats", "qdrant", "neo4j", "s3", "oss"}
}
