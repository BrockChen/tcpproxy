package tcpproxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"
)

// Proxy is a TCP server that takes an incoming request and sends it to another
// server, proxying the response back to the client.
type Proxy struct {
	// Target address
	Target *net.TCPAddr

	// Local address
	Addr *net.TCPAddr

	// Director must be a function which modifies the request into a new request
	// to be sent. Its response is then copied back to the client unmodified.
	Director func(b *[]byte)

	// If config is not nil, the proxy connects to the target address and then
	// initiates a TLS handshake.
	Config *tls.Config

	// Timeout is the duration the proxy is staying alive without activity from
	// both client and target. Also, if a pipe is closed, the proxy waits 'timeout'
	// seconds before closing the other one. By default timeout is 60 seconds.
	Timeout time.Duration
}

// NewProxy created a new proxy which sends all packet to target. The function dir
// intercept and can change the packet before sending it to the target.
func NewProxy(target *net.TCPAddr, dir func(*[]byte), config *tls.Config) *Proxy {
	p := &Proxy{
		Target:   target,
		Director: dir,
		Timeout:  time.Minute,
		Config:   config,
	}
	return p
}

// ListenAndServe listens on the TCP network address laddr and then handle packets
// on incoming connections.
func (p *Proxy) ListenAndServe(laddr *net.TCPAddr) {
	p.Addr = laddr

	var listener net.Listener
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p.serve(listener)
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it uses TLS
// protocol. Additionally, files containing a certificate and matching private key
// for the server must be provided.
func (p *Proxy) ListenAndServeTLS(laddr *net.TCPAddr, certFile, keyFile string) {
	p.Addr = laddr

	var listener net.Listener
	cer, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	listener, err = tls.Listen("tcp", laddr.String(), config)
	if err != nil {
		fmt.Println(err)
		return
	}

	p.serve(listener)
}

func (p *Proxy) serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go p.handleConn(conn)
	}
}

// handleConn handles connection.
func (p *Proxy) handleConn(conn net.Conn) {
	// connects to target server
	var rconn net.Conn
	var err error
	if p.Config == nil {
		rconn, err = net.Dial("tcp", p.Target.String())
	} else {
		rconn, err = tls.Dial("tcp", p.Target.String(), p.Config)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	// pipeDone counts closed pipe
	var pipeDone int32
	var timer *time.Timer

	// write to dst what it reads from src
	var pipe = func(src, dst net.Conn, filter func(b *[]byte)) {
		defer func() {
			// if it is the first pipe to end...
			if v := atomic.AddInt32(&pipeDone, 1); v == 1 {
				// ...wait 'timeout' seconds before closing connections
				timer = time.AfterFunc(p.Timeout, func() {
					// test if the other pipe is still alive before closing conn
					if atomic.AddInt32(&pipeDone, 1) == 2 {
						conn.Close()
						rconn.Close()
					}
				})
			} else if v == 2 {
				conn.Close()
				rconn.Close()
				timer.Stop()
			}
		}()

		buff := make([]byte, 65535)
		for {
			n, err := src.Read(buff)
			if err != nil {
				return
			}
			b := buff[:n]

			if filter != nil {
				filter(&b)
			}

			n, err = dst.Write(b)
			if err != nil {
				return
			}
		}
	}

	go pipe(conn, rconn, p.Director)
	go pipe(rconn, conn, nil)
}
