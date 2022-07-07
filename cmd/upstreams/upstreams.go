// Command upstreams runs one or more "upstreams" that
// can sit behind a load balancer, and tracks the number
// connections each has received.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
)

func main() {
	count := flag.Int("n", 1, "the number of upstreams to run")
	startPort := flag.Int("p", 8888, "the port for the first upstream to use")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	upstreams := make([]echoUpstream, *count)
	for i := 0; i < *count; i++ {
		upstreams[i] = echoUpstream{
			port: *startPort + i,
		}
		ii := i
		go upstreams[ii].run(ctx)
	}

	fmt.Printf("running %d listeners starting at port %d...\n", *count, *startPort)
	fmt.Printf("press ctrl-c to quit and print stats\n")
	<-ctx.Done()
	fmt.Println()
	for i := range upstreams {
		fmt.Printf("Upstream %d: %d connections\n", i, atomic.LoadInt64(&upstreams[i].count))
	}
}

// echoUpstream listens on a particular port and echos all input
// back to the client
type echoUpstream struct {
	count int64
	port  int

	l *net.TCPListener
}

func (e *echoUpstream) run(ctx context.Context) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", e.port))
	if err != nil {
		return err
	}

	defer l.Close()

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		atomic.AddInt64(&e.count, 1)
		go func() {
			defer conn.Close()
			if _, err := io.Copy(conn, conn); err != nil && err != io.EOF {
				log.Fatal(err)
			}
		}()
	}
}
