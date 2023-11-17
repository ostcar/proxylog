package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/ostcar/proxylog/proxy"
	"github.com/ostcar/proxylog/sizelog"
	"golang.org/x/sync/errgroup"
)

const (
	listenAddr    = ":4567"
	webserverAddr = ":9050"
)

func main() {
	ctx, cancel := interruptContext()
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	sl := new(sizelog.SizeLog)
	eg, ctx := errgroup.WithContext(ctx)

	var file io.Writer
	if len(os.Args) >= 2 {
		f, err := os.OpenFile(os.Args[1], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}

		file = f
	}

	eg.Go(func() error {
		sl.Background(ctx, file)
		return nil
	})

	eg.Go(func() error {
		return sl.Run(ctx, webserverAddr)
	})

	eg.Go(func() error {
		return proxy.Start(ctx, listenAddr, sl, nil)
	})

	return eg.Wait()
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
