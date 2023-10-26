package flopt

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRefresher_RefreshCache(t *testing.T) {
	flags := NewFlags(WithBootstrapMap(map[string]bool{"key": false}))
	fn := func(ctx context.Context, s string) (bool, error) {
		return true, nil
	}
	r, err := NewRefresher(zap.NewNop(), flags, fn)
	if err != nil {
		t.Errorf("NewRefresher() error = %v", err)
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(20*time.Millisecond))
	r.RefreshCache(ctx, time.Duration(10*time.Millisecond))
	cancelFunc()

	if flags.values["key"] != true {
		t.Errorf("RefreshCache() did not update the cache")
	}
}
