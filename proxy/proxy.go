package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/sync/errgroup"
)

// LogSizer logs the size of a package.
type LogSizer interface {
	LogSize(size int)
}

// LogSizeFunc impelements the LogSizer interface with a function.
type LogSizeFunc func(size int)

// LogSize cols LogSize.
func (ls LogSizeFunc) LogSize(size int) {
	ls(size)
}

// Start starts the proxy. Runs until an error happens or the context if done.
func Start(ctx context.Context, toAddr, listenAddr string, incomming, outgoing LogSizer) error {
	listener, err := new(net.ListenConfig).Listen(ctx, "tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("start listening on %s: %w", listenAddr, err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept connection: %w", err)
		}

		go func() {
			if err := handleConn(ctx, conn, toAddr, outgoing, incomming); err != nil {
				// TODO: Maybe use an error handler.
				log.Printf("Connection Error: %v", err)
			}
		}()

	}
}

func handleConn(ctx context.Context, conn net.Conn, toAddr string, outgoing, incomming LogSizer) error {
	defer conn.Close()

	remote, err := net.Dial("tcp", toAddr)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer remote.Close()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return copy(ctx, remote, conn, outgoing)
	})

	eg.Go(func() error {
		return copy(ctx, conn, remote, incomming)
	})

	return eg.Wait()
}

func copy(ctx context.Context, dst io.Writer, src io.Reader, logSizer LogSizer) error {
	bufSize := 32 * 1024
	buf := make([]byte, bufSize)

	for ctx.Err() == nil {
		size, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read: %w", err)
		}

		if size == 0 || ctx.Err() != nil {
			fmt.Printf("size: %v\n", size)
			continue
		}

		nw, err := dst.Write(buf[0:size])
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}

		if logSizer != nil {
			logSizer.LogSize(nw)
		}

		if size != nw {
			return io.ErrShortWrite
		}
	}

	return ctx.Err()
}
