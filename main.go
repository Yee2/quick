package main

import (
	"os"
	"fmt"
	"flag"
)
var	flags = struct {
		CA string
		CRT string
		Key string
		S bool
		C bool
		Server string
		Local string
		Address string
		KeepAlive bool
	}{}
// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	for i := range os.Args{
		if os.Args[i] == "-s"{
			Server()
			return
		}else if os.Args[i] == "-c"{
			Client()
			return
		}
	}
	Usage()
}

func Usage(){
	sHelp := flag.NewFlagSet("server mode",flag.ContinueOnError)
	sHelp.String("ca", "ca.crt", "root certificate")
	sHelp.String("crt", "server.crt", "certificate")
	sHelp.String("key", "server.key", "key")
	sHelp.String("addr", "0.0.0.0:4242", "host name or IP address of your remote server")
	sHelp.Bool("s",false,"server mode")
	sHelp.Bool("c",false,"run as a client")
	sHelp.Usage()
	fmt.Println()
	cHelp := flag.NewFlagSet("client mode",flag.ContinueOnError)
	cHelp.String("ca", "ca.crt", "root certificate")
	cHelp.String("crt", "server.crt", "certificate")
	cHelp.String("key", "server.key", "key")
	cHelp.String("addr", "", "host name or IP address of your remote server")
	cHelp.Bool("s",false,"server mode")
	cHelp.Bool("c",false,"run as a client")
	cHelp.Usage()
}