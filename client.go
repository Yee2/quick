package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"github.com/Yee2/quick/client"
	"io"
	"io/ioutil"
	"net"
	"reflect"
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
		logf("Certificate error:%s", err)
		return
	}
	manager, err := client.NewManager(flags.Server, config)
	if err != nil {
		logf("Unable to connect to remote server:%s", err)
		return
	}
	if err = Listen(manager); err != nil {
		logf("Unhandled error:%s", err)
		return
	}
}
func Listen(session *client.Manager) (e error) {
	listener, err := net.Listen("tcp", flags.Local)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {

		conn, err := listener.Accept()
		if err != nil {
			logf("error(%s):%s", reflect.TypeOf(err), err)
			continue
		}
		go func(conn net.Conn) {
			stream, err := session.NewStream()
			if err != nil {
				e = err
				//TODO: 处理错误，终止客户端运行
				//cancel()
				return
			}
			if flags.Redirect {
				err = direct(conn, stream)
			} else {
				err = tunnel(conn, stream)
			}
		}(conn)
	}
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

	cert, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      roots,
	}, nil
}

func tunnel(left io.ReadWriteCloser, right io.ReadWriteCloser) (err error) {
	defer right.Close()
	defer left.Close()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		io.Copy(left, right)
		cancel()
	}()
	go func() {
		io.Copy(right, left)
		cancel()
	}()
	<-ctx.Done()
	return
}

func die(err error) {
	if err != nil {
		panic(err)
	}
}
