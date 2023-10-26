// Package flopt provides a cache for feature flags.
package flopt

import (
	"fmt"
	"strconv"
	"sync"
)

// Flags is a cache of flag keys to their values.
type Flags struct {
	values map[string]bool
	m      sync.RWMutex
}

// NewFlags returns a new Flags cache.
func NewFlags(options ...Option) *Flags {
	f := Flags{
		values: map[string]bool{},
		m:      sync.RWMutex{},
	}
	for _, o := range options {
		o(&f)
	}
	return &f
}

// Update updates the value of the flag key.
// If the flag is not in the cache, it will be added.
func (f *Flags) Update(key string, value bool) {
	f.m.Lock()
	defer f.m.Unlock()

	f.values[key] = value
}

// BatchUpdate updates the values of the flags in the map.
// If the flag is not in the cache, it will be added.
// If the flag is in the cache, it will be updated.
// If the map is empty or nil, it will do nothing.
func (f *Flags) BatchUpdate(newFlags map[string]bool) {
	f.m.Lock()
	defer f.m.Unlock()

	// Even overwriting the map is faster, We don't want to overwrite the map, because we want to keep the existing values.
	for k, v := range newFlags {
		f.values[k] = v
	}
}

func (f *Flags) Read(key string) (value bool, found bool) {
	f.m.RLock()
	defer f.m.RUnlock()

	value, found = f.values[key]
	return
}

type Option func(*Flags)

// WithBootstrapMap is an option that allows the client to bootstrap the cache with a map of key/value pairs.
func WithBootstrapMap(bm map[string]bool) Option {
	return func(s *Flags) {
		s.m.Lock()
		defer s.m.Unlock()

		s.values = bm
	}
}

// WithBootstrapPairs is an option that allows the client to bootstrap the cache with a list of key/value pairs.
// The pairs must be even, and the value must be a valid boolean representation (true/false, 1/0, t/f, T/F, true/false, TRUE/FALSE, True/False).
// If the value is not a valid boolean representation, it will panic.
// If the pairs are not even, it will panic.
// If the pairs are empty or nil, it will do nothing.
func WithBootstrapPairs(pairs ...string) Option {
	return func(s *Flags) {
		s.m.Lock()
		defer s.m.Unlock()

		if len(pairs)%2 != 0 {
			panic("pairs must be even")
		}

		for i := 0; i < len(pairs); i = i + 2 {
			parseBool, err := strconv.ParseBool(pairs[i+1])
			if err != nil {
				panic(fmt.Sprintf("%q is not a boolean", pairs[i+1]))
			}
			s.values[pairs[i]] = parseBool
		}
	}
}

// IsEnabled returns the value of the flag key.
// If the flag is not in the cache, it will return the fallback value.
// Also, it will register the flag to be fetched when it is not in the cache.
func (f *Flags) IsEnabled(key string, fallbackValue bool) bool {
	if b, ok := f.Read(key); !ok {
		// We don't have a value for this flag, so return the fallback value.
		// This is particularly useful for when the flag is not yet in the cache. So we register the flag to be fetched
		f.Update(key, fallbackValue)
		return fallbackValue
	} else {
		return b
	}
}
