package main

import (
	"gopkg.in/urfave/cli.v2"
	"os"
	"log"
)

func main() {
	app := cli.App{
		Name:"quick",
	}
	app.Version = "201809"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name: "ca",
			Value:"ca.crt",
			Usage:"root certificate",
		},
		&cli.BoolFlag{
			Name: "debug",
			Aliases:[]string{"D"},
			Value:true,
			Usage:"debug mode",
		},
	}
	app.Before = func(ctx *cli.Context) error {
		debug = ctx.Bool("debug")
		return nil
	}
	app.Commands = []*cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Flags:   []cli.Flag{
				&cli.StringFlag{
					Name:"crt",
					Value:"server.crt",
					Usage:"server certificate",
				},
				&cli.StringFlag{
					Name:"key",
					Value:"server.key",
					Usage:"server key",
				},
				&cli.StringFlag{
					Name:"remote",
					Value:"0.0.0.0:4242",
					Usage:"host name or IP address of your remote server",
				},
				&cli.BoolFlag{
					Name:"keep-alive",
					Value:true,
					Usage:"",
				},
			},
			Action:  Server,
		},
		{
			Name:    "client",
			Aliases: []string{"c"},
			Flags:   []cli.Flag{
				&cli.StringFlag{
					Name:"crt",
					Value:"client.crt",
					Usage:"client certificate",
				},
				&cli.StringFlag{
					Name:"key",
					Value:"client.key",
					Usage:"client key",
				},
				&cli.StringFlag{
					Name:"remote",
					Value:"",
					Usage:"host name or IP address of your remote server",
				},
				&cli.StringFlag{
					Name:"local",
					Value:"0.0.0.0:1080",
					Usage:"local listening port",
				},
				&cli.BoolFlag{
					Name:"redirect",
					Value:false,
					Usage:"redirect(experiment)",
				},
			},
			Action:  Client,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
