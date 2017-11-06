package main

import "gopkg.in/urfave/cli.v1"

var (
	HTTPPort = 7410
)

func main() {
	app := cli.NewApp()
	app.Name = "Spaco cosmos AutoPoster"
	app.Usage = "Used for testing Spaco cosmos"
	app.Flags = cli.FlagsByName{
		cli.IntFlag{
			Name:        "http-port,p",
			Destination: &HTTPPort,
			Value:       HTTPPort,
		},
	}
	app.Commands = cli.Commands{
		{},
	}
	app.Action = cli.ActionFunc(func(ctx *cli.Context) error {
		return nil
	})
}
