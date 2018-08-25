package main

import (
	"net"
	"io"
	"crypto/tls"
	"flag"
	"io/ioutil"
	"crypto/x509"
	"reflect"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/qerr"
)

func Client() {
	var err error
	flag.StringVar(&flags.CA, "ca", "ca.crt", "root certificate")
	flag.StringVar(&flags.CRT, "crt", "client.crt", "client certificate")
	flag.StringVar(&flags.Key, "key", "client.key", "client key")
	flag.StringVar(&flags.Server, "addr", "", "host name or IP address of your remote server")
	flag.StringVar(&flags.Local, "local", "0.0.0.0:1080", "local listening port")
	flag.BoolVar(&flags.S, "s", false, "server mode")
	flag.BoolVar(&flags.KeepAlive, "keep", false, "keep alive")
	flag.BoolVar(&flags.Redirect, "redirect", false, "redirect")
	flag.BoolVar(&flags.C, "c", false, "run as a client")
	flag.Parse()

	config, err := createClientConfig(flags.CA, flags.CRT, flags.Key)
	if err != nil {
		logf("Unable to connect to remote server:%s", err)
		return
	}
	for {
		if err := Dial(config); err != nil {
			logf("error(%s):%s", reflect.TypeOf(err), err)
		}
	}
}
func Dial(config *tls.Config) (e error) {
	logf("start dial")
	session, err := quic.DialAddr(flags.Server, config, &quic.Config{KeepAlive: true})
	if err != nil{
		return err
	}
	logf("connect done.")
	defer session.Close()
	listener, err := net.Listen("tcp", flags.Local)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	//logf("Listening %s \n", flags.Local)
	Redial := false
	for {
		conn, err := listener.Accept()
		if err != nil {
			if Redial {
				return
			}
			logf("error(%s):%s", reflect.TypeOf(err), err)
			continue
		}
		//logf("new client:%s \n", conn.RemoteAddr())
		go func() {
			if flags.Redirect {
				err = direct(conn,session)
			} else {
				err = tunnel(conn, session)
			}
			if err != nil {
				if QuicError, ok := err.(*qerr.QuicError); ok && QuicError.ErrorCode == qerr.PublicReset {
					// 服务端重启过，客户端需要重新拨号
					e = err
					Redial = true
					listener.Close()
					return
				}
				logf("error(%s):%s", reflect.TypeOf(err), err)
			}
		}()
	}
	return nil
}
func createClientConfig(ca, crt, key string) (*tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		panic("failed to parse root certificate")
	}

	logf("CA:%s\nCRT:%s\nkey:%s", ca, crt, key)
	cert, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
	}, nil
}

func tunnel(client io.ReadWriteCloser, session quic.Session) (err error) {
	defer client.Close()
	stream, err := session.OpenStream()
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

func die(err error){
	if err != nil{
		panic(err)
	}
}