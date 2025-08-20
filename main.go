package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

var (
	sshProxyFlag = &cli.StringFlag{
		Name:    "ssh-proxy",
		Usage:   "SSH proxy",
		Sources: cli.EnvVars("SSH_PROXY"),
	}
	influxdbServerUrlFlag = &cli.StringFlag{
		Category: "InfluxDB",
		Name:     "influxdb-server-url",
		Usage:    "InfluxDB server URL",
		Sources:  cli.EnvVars("INFLUXDB_SERVER_URL"),
		Value:    "http://localhost:8086",
	}
	influxdbAuthTokenFlag = &cli.StringFlag{
		Category: "InfluxDB",
		Name:     "influxdb-auth-token",
		Usage:    "InfluxDB auth token",
		Sources:  cli.EnvVars("INFLUXDB_AUTH_TOKEN"),
		Required: true,
	}
	configFileFlag = &cli.StringFlag{
		Name:     "config-file",
		Usage:    "config file",
		Required: true,
	}
	outputFileFlag = &cli.StringFlag{
		Name:     "output-file",
		Usage:    "output file",
		Required: true,
	}
	responsiveFlag = &cli.BoolFlag{
		Name:  "responsive",
		Usage: "responsive",
		Value: true,
	}
	animationDurationFlag = &cli.DurationFlag{
		Name:  "animation-duration",
		Usage: "animation duration",
		Value: 1 * time.Second,
	}

	app = &cli.Command{
		Name:  "influxdb-tool",
		Usage: "influxdb-tool",
		Commands: []*cli.Command{
			{
				Name:  "report",
				Usage: "report",
				Flags: []cli.Flag{
					sshProxyFlag,
					influxdbServerUrlFlag,
					influxdbAuthTokenFlag,
					configFileFlag,
					outputFileFlag,
					responsiveFlag,
					animationDurationFlag,
				},
				Action: doReport,
			},
		},
	}
)

func main() {
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
