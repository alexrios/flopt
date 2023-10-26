package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/alexrios/flopt"

	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:9001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// client
	client := flipt.NewFliptClient(conn)
	// logs
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, os.Kill)

	ctx := context.Background()
	flags := flopt.NewFlags()
	fliptFetcher := flopt.NewFliptFetcher(client, "default")

	refresher, err := flopt.NewRefresher(logger, flags, fliptFetcher.Fetch)
	if err != nil {
		panic(err)
	}

	go func() {
		refresher.RefreshCache(ctx, 2*time.Second)
		exitChan <- os.Interrupt // I know this is not the ~best way~ to do this, but it's just an example
	}()

	key := "test-flag"
	for {
		select {
		case <-time.After(5 * time.Second):
			enabled := flags.IsEnabled(key, false)
			fmt.Println(key, enabled)
		case <-exitChan:
			return
		}
	}
}
