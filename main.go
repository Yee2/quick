package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"

	"github.com/lucas-clemente/quic-go"
)

const addr = "0.0.0.0:4242"


// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	if len(os.Args) < 2{
		Usage()
		os.Exit(0)
	}
	switch os.Args[1]{
	case "server","s":
		if err := Server(); err != nil {
			logf("%s", err)
		}
	case "client","c":
		if len(os.Args) < 3{
			Usage()
			os.Exit(0)
		}
		if err := Client(os.Args[2]); err != nil {
			logf("%s", err)
		}
	default:
		Usage()
	}
}

// Start a server that echos all data on the first stream opened by the client
func Server() error {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		return err
	}
	logf("listen at %s",addr)
	session, err := listener.Accept()
	if err != nil {
		return err
	}
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			panic(err)
		}
		go handleConnection(stream)
	}
	return nil
}


// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

func Usage(){
	logf("用法:%s server|client [address]",os.Args[0])
}