// +build windows

package main

import (
	"net"
	"github.com/lucas-clemente/quic-go"
)

func direct(conn net.Conn,_ quic.Session) error {
	defer conn.Close()
	panic(errors.New("does not support Windows platform"))
}