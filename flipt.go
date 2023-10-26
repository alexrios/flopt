package flopt

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/rpc/flipt"
)

// FliptFetcher is a flag fetcher using the flipt client.
type FliptFetcher struct {
	client    flipt.FliptClient
	namespace string // flipt namespace
}

// NewFliptFetcher returns a new Fetcher.
func NewFliptFetcher(c flipt.FliptClient, ns string) FliptFetcher {
	return FliptFetcher{
		client:    c,
		namespace: ns,
	}
}

// Fetch fetches the flag value for the given key.
// When the flag is not found, it returns false.
// When there is an error fetching the flag, it returns an error.
// Otherwise, it returns the fetched flag value.
func (f FliptFetcher) Fetch(ctx context.Context, key string) (bool, error) {
	flagRequest := flipt.GetFlagRequest{
		Key:          key,
		NamespaceKey: f.namespace,
	}
	flag, err := f.client.GetFlag(ctx, &flagRequest)
	if err != nil {
		return false, fmt.Errorf("error fetching flag: %w", err)
	}
	if flag == nil {
		return false, fmt.Errorf("flag %s not found", key)
	}
	return flag.Enabled, nil
}
