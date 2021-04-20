package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	collector "github.com/jesusfcr/sqs-prometheus-exporter/pkg/collector"

	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	port                  = getEnv("PORT", "9434")
	intervalStr           = getEnv("INTERVAL", "1")
	endpoint              = getEnv("ENDPOINT", "metrics")
	keepRunningOnErrorStr = getEnv("KEEP_RUNNING", "true")
	sqsNamePrefix         = getEnv("SQS_QUEUE_NAME_PREFIX", "")
	sqsEndpoint           = getEnv("AWS_SQS_ENDPOINT", "")
)

func main() {
	ctx := context.Background()
	interval, conversionError := strconv.ParseUint(intervalStr, 10, 64)
	if conversionError != nil {
		panic(conversionError)
	}

	httpServer, err := setupMetricsServer()
	if err != nil {
		fmt.Println(err)
		return
	}

	errChanel := make(chan error)

	go func() {
		e := httpServer.ListenAndServe()
		errChanel <- e
	}()

	scheduler := gocron.NewScheduler(time.UTC)

	scheduler.Every(interval).Minutes().Do(startMonitoring, errChanel)
	scheduler.Start()

	fmt.Println(fmt.Sprintf("Metrics server listening at port %v with monitoring interval of %v minute(s).", httpServer.Addr, interval))

	keepRunningOnError, _ := strconv.ParseBool(keepRunningOnErrorStr)
	if keepRunningOnError {
		for {
			err = <-errChanel
			fmt.Println(err)

			index := strings.Index(err.Error(), "[MONITORING ERROR]")
			if index == -1 {
				break
			}
		}
	} else {
		err = <-errChanel
		fmt.Println(err)
	}

	fmt.Println("Terminating the server and monitoring")
	httpServer.Shutdown(ctx)
	scheduler.Clear()
}

func startMonitoring(errChanel chan error) {
	err := collector.MonitorSQS(sqsNamePrefix, sqsEndpoint)
	if err != nil {
		errChanel <- err
	}
	return
}

func setupMetricsServer() (*http.Server, error) {
	var (
		listenAddress = flag.String("web.listen-address", ":"+port, "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/"+endpoint, "Path under which to expose metrics.")
	)
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
	    	<html>
	        <head><title>SQS Prometheus Exporter</title></head>
	        <body>
	        <h1>SQS Prometheus Exporter</h1>
	        <p><a href='` + *metricsPath + `'>Metrics</a></p>
	        </body>
	        </html>`))
	})
	mux.Handle(*metricsPath, promhttp.Handler())
	httpServer := &http.Server{
		Addr:    *listenAddress,
		Handler: mux,
	}

	return httpServer, nil
}

// GetEnv returns the value of an environment variable with a fallback
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
