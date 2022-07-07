// Command client is used to test the behavior of a TCP
// load balancer. It assumes that the load balancer fronts
// one or more TCP echo servers, like the ones found in
// commmand upstreams.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
)

func main() {
	port := flag.Int("p", 0, "port to connect to")
	oneshot := flag.Bool("oneshot", false, "make a single connection and verify that the output is correct")
	// TODO:
	// number of worker routines
	// timing info (hold connections open longer)
	flag.Parse()

	if *port == 0 {
		log.Fatal("missing required port argument")
	}

	if *oneshot {
		if err := runOnce(*port); err != nil {
			log.Fatal(err)
		}
		return
	}

	// TODO: run a bunch of times..
}

// runOnce writes a random number of arbitrary data to the
// specified TCP port and verifies that the same data is
// echoed back
func runOnce(port int) error {
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer conn.Close()

	data := make([]byte, 128+rand.Intn(1024))
	rand.Read(data)

	n, err := conn.Write(data)
	if err != nil {
		return err
	}

	if n != len(data) {
		return fmt.Errorf("could not write all %d bytes", len(data))
	}

	recv := make([]byte, n)
	if _, err := conn.Read(recv); err != nil {
		return err
	}

	if !bytes.Equal(data, recv) {
		return errors.New("data did not match")
	}

	return nil
}
