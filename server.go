package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"reflect"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/qerr"
	"gopkg.in/urfave/cli.v2"
)

func Server(ctx *cli.Context)(e error) {

	config, err := createServerConfig(ctx.String("ca"), ctx.String("crt"), ctx.String("key"))
	if err != nil {
		logf("error:%s", err)
		return err
	}
	listener, err := quic.ListenAddr(ctx.String("remote"), config, &quic.Config{KeepAlive: ctx.Bool("keep-alive")})
	if err != nil {
		logf("error:%s", err)
		return err
	}
	defer listener.Close()
	logf("listen at %s", ctx.String("remote"))
	for {
		session, err := listener.Accept()
		if err != nil {
			logf("error(%s):%s", reflect.TypeOf(err), err)
			continue
		}
		logf("new session from %s", session.RemoteAddr())
		go severSessionHandle(session)
	}
	return nil
}

func severSessionHandle(session quic.Session) {
	defer session.Close()
	for {
		stream, err := session.AcceptStream()
		if err == nil {
			go handleConnection(stream)
			continue
		}
		if QuicError, ok := err.(*qerr.QuicError); ok && QuicError.Timeout() {
			logf("session timeout, closing session")
			break
		} else if ok && QuicError.ErrorCode == qerr.PeerGoingAway {
			logf("session closed by client")
			break
		}
		logf("error(%s):%s", reflect.TypeOf(err), err)
	}
}

func createServerConfig(ca, crt, key string) (*tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		panic("failed to parse root certificate")
	}

	cert, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    roots,
	}, nil
}
