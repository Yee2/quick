package main

import (
	"flag"
	"crypto/tls"
	"io/ioutil"
	"crypto/x509"
	"reflect"

	"github.com/lucas-clemente/quic-go/qerr"
	"github.com/lucas-clemente/quic-go"
)

func Server()  {
	flag.StringVar(&flags.CA,"ca", "ca.crt", "root certificate")
	flag.StringVar(&flags.CRT,"crt", "server.crt", "server certificate")
	flag.StringVar(&flags.Key,"key", "server.key", "server key")
	flag.StringVar(&flags.Address,"addr", "0.0.0.0:4242", "host name or IP address of your remote server")
	flag.BoolVar(&flags.S,"s",false,"server mode")
	flag.BoolVar(&flags.KeepAlive,"keep",false,"keep alive")
	flag.BoolVar(&flags.C,"c",false,"run as a client")
	flag.Parse()
	config,err := createServerConfig(flags.CA,flags.CRT,flags.Key)
	if err != nil{
		logf("error:%s",err)
		return
	}
	listener, err := quic.ListenAddr(flags.Address, config, &quic.Config{KeepAlive:flags.KeepAlive})
	if err != nil {
		logf("error:%s",err)
		return
	}
	defer listener.Close()
	logf("listen at %s",flags.Address)
	for {
		session, err := listener.Accept()
		if err != nil {
			logf("error(%s):%s",reflect.TypeOf(err),err)
			continue
		}
		logf("new session from %s",session.RemoteAddr())
		go func() {
			defer session.Close()
			for {
				stream, err := session.AcceptStream()
				if err != nil {
					if QuicError,ok := err.(*qerr.QuicError);ok && QuicError.Timeout(){
						logf("session closed by client")
						break
					}
					if QuicError,ok := err.(*qerr.QuicError);ok && QuicError.ErrorCode == qerr.PeerGoingAway{
						logf("session closed by client")
						break
					}
					logf("error(%s):%s",reflect.TypeOf(err),err)
					continue
				}
				go handleConnection(stream)
			}
		}()
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
