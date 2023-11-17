package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/ostcar/proxylog/proxy"
)

const listenAddr = ":4567"

func main() {
	ctx, cancel := interruptContext()
	defer cancel()

	if err := run(ctx, os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("Usage: %s SERVER_ADDR", os.Args[0])
	}

	url, err := url.Parse(args[1])
	if err != nil {
		return fmt.Errorf("invalid serer addr: %w", err)
	}

	browserAddr := listenAddr
	if strings.HasPrefix(listenAddr, ":") {
		browserAddr = "localhost" + listenAddr
	}

	fmt.Printf("Start browser and call %s://%s\n", url.Scheme, browserAddr)

	host := url.Hostname() + ":" + url.Port()
	if url.Port() == "" {
		port := "80"
		if url.Scheme == "https" {
			port = "443"
		}
		host += port
	}

	return proxy.Start(ctx, host, listenAddr, proxy.LogSizeFunc(func(size int) { fmt.Println(size) }), nil)
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
