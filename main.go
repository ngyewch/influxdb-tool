package main

import (
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	sshProxyFlag = &cli.StringFlag{
		Name:    "ssh-proxy",
		Usage:   "SSH proxy",
		EnvVars: []string{"SSH_PROXY"},
	}
	influxdbServerUrlFlag = &cli.StringFlag{
		Category: "InfluxDB",
		Name:     "influxdb-server-url",
		Usage:    "InfluxDB server URL",
		EnvVars:  []string{"INFLUXDB_SERVER_URL"},
		Value:    "http://localhost:8086",
	}
	influxdbAuthTokenFlag = &cli.StringFlag{
		Category: "InfluxDB",
		Name:     "influxdb-auth-token",
		Usage:    "InfluxDB auth token",
		EnvVars:  []string{"INFLUXDB_AUTH_TOKEN"},
		Required: true,
	}
	configFileFlag = &cli.PathFlag{
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

	app = &cli.App{
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
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
