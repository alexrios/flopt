package flopt

import (
	"testing"
)

func TestFlags_EmptyIsEnabled(t *testing.T) {
	tests := []struct {
		name          string
		flags         *Flags
		key           string
		fallbackValue bool
		want          bool
	}{
		{
			name:          "should return the fallback option when the flags are empty",
			flags:         NewFlags(),
			key:           "key",
			fallbackValue: true,
			want:          true,
		},
		{
			name:          "should return the fallback option when the flags are empty",
			flags:         NewFlags(),
			key:           "key",
			fallbackValue: false,
			want:          false,
		},
		{
			name: "should return true when flag is found (WithBootstrapMap)",
			flags: NewFlags(WithBootstrapMap(map[string]bool{
				"key": true,
			})),
			key:  "key",
			want: true,
		},
		{
			name: "should return false when flag is not found (WithBootstrapMap)",
			flags: NewFlags(WithBootstrapMap(map[string]bool{
				"key": false,
			})),
			key:  "key",
			want: false,
		},
		{
			name:  "should return true when flag is found (WithBootstrapPairs)",
			flags: NewFlags(WithBootstrapPairs("key", "true")),
			key:   "key",
			want:  true,
		},
		{
			name:  "should return false when flag is not found (WithBootstrapPairs)",
			flags: NewFlags(WithBootstrapPairs("key", "false")),
			key:   "key",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.flags
			if got := f.IsEnabled(tt.key, tt.fallbackValue); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlags_Read(t *testing.T) {
	tests := []struct {
		name  string
		flags *Flags
		key   string
		want  bool
	}{
		{
			name: "should return true when reading a true flag",
			flags: NewFlags(WithBootstrapMap(map[string]bool{
				"key-read": true,
			})),
			key:  "key-read",
			want: true,
		},
		{
			name: "should return false when reading a false flag",
			flags: NewFlags(WithBootstrapMap(map[string]bool{
				"key-read": false,
			})),
			key:  "key-read",
			want: false,
		},
		{
			name:  "should return false when reading fail",
			flags: NewFlags(),
			key:   "key-read",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.flags
			got, _ := f.Read(tt.key)
			if got != tt.want {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestFlags_Update(t *testing.T) {
	tests := []struct {
		name  string
		flags *Flags
		key   string
		value bool
	}{
		{
			name:  "should set the flag to true",
			flags: NewFlags(),
			key:   "key1",
			value: true,
		},
		{
			name: "should set the flag to true",
			flags: NewFlags(WithBootstrapMap(map[string]bool{
				"key1": true,
			})),
			key:   "key1",
			value: true,
		},
		{
			name:  "should set the flag to true",
			flags: NewFlags(WithBootstrapPairs("key1", "true")),
			key:   "key1",
			value: true,
		},
		{
			name:  "should set the flag to false",
			flags: NewFlags(),
			key:   "key1",
			value: false,
		},
		{
			name: "should set the flag to false",
			flags: NewFlags(WithBootstrapMap(map[string]bool{
				"key1": false,
			})),
			key:   "key1",
			value: false,
		},
		{
			name:  "should set the flag to false",
			flags: NewFlags(WithBootstrapPairs("key1", "false")),
			key:   "key1",
			value: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.flags
			f.Update(tt.key, tt.value)
			value, found := f.Read(tt.key)
			if !found {
				t.Errorf("Key %s not found after the update", tt.key)
			}
			if value != tt.value {
				t.Errorf("Key %s has value %v, want %v", tt.key, value, tt.value)
			}
		})
	}
}

func TestFlags_BatchUpdate(t *testing.T) {
	flags := NewFlags()
	newFlags := map[string]bool{
		"key1": true,
		"key2": false,
	}
	flags.BatchUpdate(newFlags)
	for k, _ := range flags.values {
		value, found := flags.Read(k)
		if !found {
			t.Errorf("Key %s not found after the update", k)
		}
		if newFlags[k] != value {
			t.Errorf("Key %s has value %v, want %v", k, newFlags[k], value)
		}
	}
}
