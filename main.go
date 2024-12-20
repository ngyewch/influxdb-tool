package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
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
