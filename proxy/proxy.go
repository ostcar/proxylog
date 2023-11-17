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

// Start starts the proxy. Runs until an error happens or the context if done.
func Start(ctx context.Context, listenAddr string, incomming, outgoing LogSizer) error {
	fmt.Printf("Listen socks4 proxy on  %s\n", listenAddr)
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
			if err := handleConn(ctx, conn, outgoing, incomming); err != nil {
				// TODO: Maybe use an error handler.
				log.Printf("Connection Error: %v", err)
			}
		}()

	}
}

func handleConn(ctx context.Context, conn net.Conn, outgoing, incomming LogSizer) error {
	defer conn.Close()

	addr, err := socks4connect(conn)
	if err != nil {
		return fmt.Errorf("read socks4 header: %w", err)
	}

	remote, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer remote.Close()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := copy(ctx, remote, conn, outgoing); err != nil {
			return fmt.Errorf("copy from conn to remote: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		if err := copy(ctx, conn, remote, incomming); err != nil {
			return fmt.Errorf("copy from remote to conn: %w", err)
		}

		return nil
	})

	return eg.Wait()
}

func socks4connect(rw io.ReadWriter) (string, error) {
	buf := make([]byte, 9)
	if _, err := io.ReadAtLeast(rw, buf, 9); err != nil {
		return "", fmt.Errorf("read : %w", err)
	}

	if buf[0] != 4 || buf[1] != 1 {
		return "", fmt.Errorf("expecting a socks4 new tcp packe")
	}

	port := int(buf[2])*256 + int(buf[3])

	ip := net.IP(buf[4:8])
	addr := fmt.Sprintf("%s:%d", ip, port)

	if _, err := rw.Write([]byte{0, 0x5A, 0, 0, 0, 0, 0, 0}); err != nil {
		return "", fmt.Errorf("writing socks4 header: %w", err)
	}

	return addr, nil
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
			logSizer.LogSize(size)
		}

		if size != nw {
			return io.ErrShortWrite
		}
	}

	return ctx.Err()
}
