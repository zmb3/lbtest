// Command client is used to test the behavior of a TCP
// load balancer. It assumes that the load balancer fronts
// one or more TCP echo servers, like the ones found in
// commmand upstreams.
package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
)

func main() {
	port := flag.Int("p", 0, "port to connect to")
	cert := flag.String("cert", "", "path to client certificate (leave blank to disable TLS)")
	key := flag.String("key", "", "path to private key (leave blank to disable TLS)")
	ca := flag.String("ca", "", "path to load balancer's certificate authority")
	oneshot := flag.Bool("oneshot", false, "make a single connection and verify that the output is correct")
	// TODO:
	// number of worker routines
	// timing info (hold connections open longer)
	flag.Parse()

	if *port == 0 {
		log.Fatal("missing required port argument")
	}

	hasCert := len(*cert) > 0
	hasKey := len(*key) > 0

	if hasCert != hasKey {
		log.Fatal("to enable TLS, specify both -cert and -key")
	}

	if *oneshot {
		if err := runOnce(*port, *cert, *key, *ca); err != nil {
			log.Fatal(err)
		}
		return
	}

	// TODO: run a bunch of times..
}

func dial(port int, cert, key, ca string) (conn net.Conn, err error) {
	useTLS := len(cert) > 0 && len(key) > 0
	if useTLS {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}

		var pool *x509.CertPool
		if len(ca) > 0 {
			caCert, err := os.ReadFile(ca)
			if err != nil {
				return nil, err
			}
			pool = x509.NewCertPool()
			pool.AppendCertsFromPEM(caCert)
		} else {
			pool, err = x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
		}

		cfg := &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{tlsCert},
		}
		tconn, err := tls.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port), cfg)
		if err != nil {
			return nil, err
		}
		return tconn, nil
	}

	conn, err = net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// runOnce writes a random number of arbitrary data to the
// specified TCP port and verifies that the same data is
// echoed back
func runOnce(port int, cert, key, ca string) error {
	conn, err := dial(port, cert, key, ca)
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
