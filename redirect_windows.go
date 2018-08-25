// +build windows

package main

import (
	"net"
	"errors"
	"github.com/lucas-clemente/quic-go"
)

func direct(conn net.Conn,_ quic.Session) error {
	defer conn.Close()
	return errors.New("does not support Windows platform")
}