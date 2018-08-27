// +build windows

package main

import (
	"github.com/lucas-clemente/quic-go"
	"net"
)

func direct(conn net.Conn, _ quic.Session) error {
	defer conn.Close()
	panic(errors.New("does not support Windows platform"))
}
