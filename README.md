# kahlys/tcpproxy

[![godoc](https://godoc.org/github.com/kahlys/tcpproxy?status.svg)](https://godoc.org/github.com/kahlys/tcpproxy) 
[![build](https://api.travis-ci.org/kahlys/tcpproxy.svg?branch=master)](https://travis-ci.org/kahlys/tcpproxy)
[![go report](https://goreportcard.com/badge/github.com/kahlys/tcpproxy)](https://goreportcard.com/report/github.com/kahlys/tcpproxy)

Simple tcp proxy package and executable binary in Golang. The executable provides both TCP and TCP/TLS connection.

## Installation

With a correctly configured [Go toolchain](https://golang.org/doc/install):
```
go get -u github.com/kahlys/tcpproxy/cmd/tcpproxy
```

## Usage

By default, the proxy address is *localhost:4444* and the target address is *localhost:80*.
```
$ tcpproxy
```
You can specify some options.
```
$ tcpproxy -h
Usage of tcpproxy:

  -laddr string
    	proxy local address (default ":4444")

  -lcert string
    	proxy certificate x509 file for tls/ssl use

  -lkey string
    	proxy key x509 file for tls/ssl use
      
  -ltls
    	tls/ssl between client and proxy
      
  -raddr string
    	proxy remote address (default ":80")
      
  -rtls
    	tls/ssl between proxy and target
      
  -t int
    	wait  seconds before closing second pipe
```
