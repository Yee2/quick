package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/Yee2/quick/client"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"gopkg.in/urfave/cli.v2"
	"errors"
)

func Client(ctx *cli.Context) (err error) {
	if ctx.String("remote") ==""{
		err = errors.New("remote is required")
		logf("%s", err)
		return
	}
	config, err := createClientConfig(ctx.String("ca"), ctx.String("crt"), ctx.String("key"))
	if err != nil {
		logf("Certificate error:%s", err)
		return
	}
	manager, err := client.NewManager(ctx.String("remote"), config)
	if err != nil {
		logf("Unable to connect to remote server:%s", err)
		return
	}
	logf("Successfully connected to a remote server:%s",ctx.String("remote"))
	if err = Listen(ctx, manager); err != nil {
		logf("Unhandled error:%s", err)
		return
	}
	return nil
}
func Listen(ctx *cli.Context, session *client.Manager) (e error) {
	listener, err := net.Listen("tcp", ctx.String("local"))
	if err != nil {
		return err
	}
	logf("Start listening:%s",ctx.String("local"))
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
			if ctx.Bool("redirect") {
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
