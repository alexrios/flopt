# Flopt

This library is an overly opinionated custom implementation of a Go gRPC client for [Flipt](https://www.flipt.io/).
The main feature is the capability to auto-refresh flags from the server and keep a local cache of the flags.

The main advantage of this approach is that flag evaluation is done locally, faster than requesting the server every time, and safer to do in hot paths.

## Main components
- Flags: A local cache of the flags, to be refreshed periodically.
- FliptFetcher: A gRPC client to fetch flags from the server.
- Refresher: A component that periodically refreshes the flags from the server using the fetcher.

## Simple usage
```go
import "github.com/alexrios/flopt"

...

ctx := context.Background()
flags := flopt.NewFlags()
fliptFetcher := flopt.NewFliptFetcher(client, "<your-flipt-namespace-here>")
refresher, _ := flopt.NewRefresher(logger, flags, fliptFetcher.Fetch)

// Keep the flags updated every 30 seconds in the background
go func () {
    refresher.RefreshCache(ctx, 30 * time.Second)
}()

// evaluate a flag
enabled := flags.IsEnabled(key, false)
fmt.Println(key, enabled)
```

### Preheating the cache
There are two main ways to preheat the cache callind NewFlags() factory:
- WithBootstrapPairs: Pass a list of key-value pairs to be used as the initial cache.
- WithBootstrapMap: Pass a map of key-value pairs to be used as the initial cache.

### Refresher options
The refresher can be configured with the following options:
- `WithRegisterer`: Pass a prometheus registerer to register metrics.
- `WithFailedAttemptsMaxDuration`: Pass a duration to be used as the max duration for failed attempts.
- `WithFailedAttemptsMaxCount`: Pass a count to be used as the max count for failed attempts.
- `WithFailedAttemptsHook`: Pass a function to be called when the number of failed attempts exceeds the max count (WithFailedAttemptsMaxCount) or the max duration (WithFailedAttemptsMaxDuration).

## Examples
- `auto_refreshing.go`: Demonstrates a connection to a Flipt server, flag refreshing, and checking flag status.

### How to Run
1. Ensure you have Go installed.
2. Clone the repository.
3. go mod tidy
4. Run `go run auto_refreshing.go` to execute the auto-refreshing example.

## Dependencies
- Flipt gRPC client: `go.flipt.io/flipt/rpc/flipt`
- gRPC: `google.golang.org/grpc`
- Zap Logging: `go.uber.org/zap`
