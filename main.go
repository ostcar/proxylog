package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ostcar/proxylog/proxy"
)

const listenAddr = ":4567"

func main() {
	ctx, cancel := interruptContext()
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	fmt.Printf("Listen socks4 proxy on  %s\n", listenAddr)

	return proxy.Start(ctx, listenAddr, proxy.LogSizeFunc(func(size int) { fmt.Println(size) }), nil)
}

// interruptContext works like signal.NotifyContext
//
// In only listens on os.Interrupt. If the signal is received two times,
// os.Exit(1) is called.
func interruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		cancel()

		// If the signal was send for the second time, make a hard cut.
		<-sigint
		os.Exit(1)
	}()
	return ctx, cancel
}
