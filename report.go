package main

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/ngyewch/go-chartjs"
	sshhelper "github.com/ngyewch/go-ssh-helper"
	"github.com/ngyewch/influxdb-tool/flux"
	"github.com/ngyewch/influxdb-tool/resources"
	"github.com/urfave/cli/v2"
	"go.octolab.org/pointer"
	"golang.org/x/crypto/ssh"
	"html/template"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

type ReportConfig struct {
	Org                string            `koanf:"org"`
	Bucket             string            `koanf:"bucket"`
	StartTime          string            `koanf:"startTime"`
	EndTime            string            `koanf:"endTime"`
	Timezone           string            `koanf:"timezone"`
	TimeDisplayFormats map[string]string `koanf:"timeDisplayFormats"`
	TimeTooltipFormat  string            `koanf:"timeTooltipFormat"`
	AggregateWindow    string            `koanf:"aggregateWindow"`
	TagOrder           []string          `koanf:"tagOrder"`
	Aliases            map[string]string `koanf:"aliases"`
	Tags               map[string]string `koanf:"tags"`
	Queries            []QueryConfig     `koanf:"queries"`
}

type QueryConfig struct {
	Tags map[string]string `koanf:"tags"`
}

type ReportData struct {
	Rows []*ReportRow
}

type ReportRow struct {
	Title      string
	MeanChart  *chartjs.LineChartConfiguration
	CountChart *chartjs.LineChartConfiguration
}

func doReport(cCtx *cli.Context) error {
	serverUrl := influxdbServerUrlFlag.Get(cCtx)
	authToken := influxdbAuthTokenFlag.Get(cCtx)
	sshProxy := sshProxyFlag.Get(cCtx)
	configFile := configFileFlag.Get(cCtx)
	outputFile := outputFileFlag.Get(cCtx)
	responsive := responsiveFlag.Get(cCtx)
	animationDuration := animationDurationFlag.Get(cCtx)

	var config ReportConfig
	err := loadConfig(configFile, &config)
	if err != nil {
		return err
	}

	options := influxdb2.DefaultOptions()
	if sshProxy != "" {
		sshClientFactory := sshhelper.DefaultSSHClientFactory()
		sshClient, err := sshClientFactory.CreateForAlias(sshProxy)
		if err != nil {
			return err
		}
		defer func(sshClient *ssh.Client) {
			_ = sshClient.Close()
		}(sshClient)

		options.SetHTTPClient(&http.Client{
			Transport: &http.Transport{
				DialContext: sshClient.DialContext,
			},
		})
	}

	client := influxdb2.NewClientWithOptions(serverUrl, authToken, options)
	defer client.Close()

	qc := &QueryClient{
		Config:            &config,
		Client:            client,
		Responsive:        responsive,
		AnimationDuration: animationDuration,
	}

	var reportData ReportData

	for _, query := range config.Queries {
		reportRow, err := qc.GenerateChart(cCtx.Context, query)
		if err != nil {
			return err
		}
		reportData.Rows = append(reportData.Rows, reportRow)
	}

	templates, err := template.New("templates").
		ParseFS(resources.TemplateFS, "templates/*.gohtml")
	if err != nil {
		return err
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	err = templates.ExecuteTemplate(f, "report.gohtml", reportData)
	if err != nil {
		return err
	}

	return nil
}

type QueryClient struct {
	Client            influxdb2.Client
	Config            *ReportConfig
	Responsive        bool
	AnimationDuration time.Duration
}

func (qc *QueryClient) GenerateChart(ctx context.Context, query QueryConfig) (*ReportRow, error) {
	var parts []string
	tags := make(map[string]string)
	for key, value := range qc.Config.Tags {
		tags[key] = value
	}
	for key, value := range query.Tags {
		tags[key] = value
	}
	for _, key := range qc.Config.TagOrder {
		value := tags[key]
		if value != "" {
			parts = append(parts, value)
		} else {
			parts = append(parts, "*")
		}
	}
	for key, value := range tags {
		if !slices.Contains(qc.Config.TagOrder, key) {
			if value != "" {
				parts = append(parts, key+" = "+value)
			} else {
				parts = append(parts, key+" = *")
			}
		}
	}

	meanChart, err := qc.doGenerateChart(ctx, query, "mean", false, false)
	if err != nil {
		return nil, err
	}
	countChart, err := qc.doGenerateChart(ctx, query, "count", false, true)
	if err != nil {
		return nil, err
	}
	return &ReportRow{
		Title:      fmt.Sprintf("%s (%s)", strings.Join(parts, " | "), qc.Config.AggregateWindow),
		MeanChart:  meanChart,
		CountChart: countChart,
	}, nil
}

func (qc *QueryClient) doGenerateChart(ctx context.Context, query QueryConfig, aggregateFn string, createEmpty bool, beginAtZero bool) (*chartjs.LineChartConfiguration, error) {
	aggregateWindow, err := time.ParseDuration(qc.Config.AggregateWindow)
	if err != nil {
		return nil, err
	}

	var tz *time.Location
	if qc.Config.Timezone != "" {
		tz1, err := time.LoadLocation(qc.Config.Timezone)
		if err != nil {
			return nil, err
		}
		tz = tz1
	}

	queryBuilder := flux.NewBuilder(qc.Config.Bucket).
		Range(qc.Config.StartTime, qc.Config.EndTime)

	tags := make(map[string]string)
	for key, value := range qc.Config.Tags {
		alias, ok := qc.Config.Aliases[key]
		if ok {
			tags[alias] = value
		} else {
			tags[key] = value
		}
	}
	for key, value := range query.Tags {
		alias, ok := qc.Config.Aliases[key]
		if ok {
			tags[alias] = value
		} else {
			tags[key] = value
		}
	}
	for key, value := range tags {
		if value != "" {
			queryBuilder = queryBuilder.Filter(fmt.Sprintf(`r["%s"] == "%s"`, key, value))
		}
	}
	queryBuilder = queryBuilder.AggregateWindow(qc.Config.AggregateWindow, aggregateFn, createEmpty)

	queryAPI := qc.Client.QueryAPI(qc.Config.Org)
	queryTableResult, err := queryAPI.Query(ctx, queryBuilder.String())
	if err != nil {
		return nil, err
	}
	defer func(queryTableResult *api.QueryTableResult) {
		_ = queryTableResult.Close()
	}(queryTableResult)

	var datasets []*chartjs.LineChartDataset
	var startTime time.Time
	var stopTime time.Time
	var lastTimestamp int64
	var currentDataset *chartjs.LineChartDataset
	tableNo := -1

	reverseAlias := make(map[string]string)
	{
		for key, value := range qc.Config.Aliases {
			reverseAlias[value] = key
		}
	}

	for queryTableResult.Next() {
		fluxRecord := queryTableResult.Record()
		if startTime.IsZero() {
			startTime = fluxRecord.Start()
		}
		if stopTime.IsZero() {
			stopTime = fluxRecord.Stop()
		}

		if tableNo != fluxRecord.Table() {
			tableNo = fluxRecord.Table()
			var parts []string
			for _, tag := range qc.Config.TagOrder {
				alias, ok := qc.Config.Aliases[tag]
				if ok {
					parts = append(parts, toString(fluxRecord.ValueByKey(alias)))
				} else {
					parts = append(parts, toString(fluxRecord.ValueByKey(tag)))
				}
			}
			var remainingKeys []string
			for key, _ := range fluxRecord.Values() {
				if (key == "_start") || (key == "_stop") || (key == "_time") || (key == "_value") || (key == "result") || (key == "table") {
					continue
				}
				alias, ok := reverseAlias[key]
				if ok {
					if !slices.Contains(qc.Config.TagOrder, alias) {
						remainingKeys = append(remainingKeys, key)
					}
				} else {
					if !slices.Contains(qc.Config.TagOrder, key) {
						remainingKeys = append(remainingKeys, key)
					}
				}
			}
			slices.Sort(remainingKeys)
			for _, key := range remainingKeys {
				parts = append(parts, toString(fluxRecord.ValueByKey(key)))
			}
			series := strings.Join(parts, " | ")
			lastTimestamp = fluxRecord.Start().UnixMilli()
			currentDataset = &chartjs.LineChartDataset{
				ControllerDatasetOptions: &chartjs.ControllerDatasetOptions{
					Label: series,
				},
			}
			datasets = append(datasets, currentDataset)
		}
		t := fluxRecord.Time()
		if tz != nil {
			t = t.In(tz)
		}
		var v *float64
		switch n := fluxRecord.Value().(type) {
		case float64:
			v = pointer.ToFloat64(n)
		case int64:
			v = pointer.ToFloat64(float64(n))
		}
		if currentDataset != nil {
			if len(currentDataset.Data) > 0 {
				for t1 := float64(lastTimestamp + aggregateWindow.Milliseconds()); t1 < float64(t.UnixMilli()); t1 += float64(aggregateWindow.Milliseconds()) {
					currentDataset.Data = append(currentDataset.Data, chartjs.Point{
						X: t1,
						Y: nil,
					})
				}
			}
			currentDataset.Data = append(currentDataset.Data, chartjs.Point{
				X: float64(t.UnixMilli()),
				Y: v,
			})
			lastTimestamp = t.UnixMilli()
		}
	}

	if len(datasets) == 0 {
		return nil, nil
	}

	return &chartjs.LineChartConfiguration{
		Data: &chartjs.LineChartData{
			Datasets: datasets,
		},
		Options: &chartjs.LineControllerChartOptions{
			CoreChartOptions: &chartjs.CoreChartOptions{
				Animation: &chartjs.AnimationSpec{
					Duration: pointer.ToFloat64(float64(qc.AnimationDuration.Milliseconds())),
				},
				Responsive:          pointer.ToBool(qc.Responsive),
				MaintainAspectRatio: pointer.ToBool(false),
			},
			Scales: map[string]chartjs.ICartesianScaleType{
				"x": &chartjs.TimeScaleOptions{
					Min: chartjs.Float64(startTime.UnixMilli()),
					Max: chartjs.Float64(stopTime.UnixMilli()),
					CartesianScaleOptions: &chartjs.CartesianScaleOptions{
						Title: &chartjs.CartesianScaleTitle{
							Display: pointer.ToBool(true),
							Text:    "Time",
						},
					},
					Time: &chartjs.TimeScaleTimeOptions{
						DisplayFormats: qc.Config.TimeDisplayFormats,
						TooltipFormat:  qc.Config.TimeTooltipFormat,
					},
				},
				"y": &chartjs.LinearScaleOptions{
					CartesianScaleOptions: &chartjs.CartesianScaleOptions{
						Title: &chartjs.CartesianScaleTitle{
							Display: pointer.ToBool(true),
							Text:    aggregateFn,
						},
					},
					BeginAtZero: pointer.ToBool(beginAtZero),
				},
			},
		},
	}, nil
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}
