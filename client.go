package main

import (
	"net"
	"io"
	"crypto/tls"

	"github.com/lucas-clemente/quic-go"
)

var (
	local   = "0.0.0.0:1080"
	session quic.Session
)

func Client(address string) (error) {
	var err error
	session, err = quic.DialAddr(address, &tls.Config{InsecureSkipVerify: true}, nil)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", local)
	if err != nil {
		return err
	}
	logf("Listening %s \n", local)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logf("%s", err)
			continue
		}
		logf("new client:%s \n", conn.RemoteAddr())
		go func() {
			err := tunnel(conn)
			if err != nil {
				logf("%s", err)
			}
		}()
	}
}

func tunnel(client io.ReadWriteCloser) (err error) {
	defer client.Close()
	stream, err := session.OpenStreamSync()
	if err != nil {
		return err
	}
	defer stream.Close()
	go func() {
		_, e := io.Copy(client, stream)
		if e != nil {
			err = e
		}
	}()
	_, e := io.Copy(stream, client)
	if e != nil {
		err = e
	}
	return
}
