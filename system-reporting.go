package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"

	"solow.xyz/system-reporting/config"
	"solow.xyz/system-reporting/influx"
)

var (
	hostname string
)

func recordMeasurement() string {
	lb := influx.NewLineBuilder("system_stats")
	lb.AddTag("device", hostname)

	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		log.Error("Error reading temperature sensors", "err", err)
	}
	for i := range temps {
		key := fmt.Sprintf("temp%d", i)
		lb.Add(key, strconv.FormatFloat(temps[i].Temperature, 'f', 1, 64))
	}

	load, err := load.Avg()
	if err != nil {
		log.Error("Error reading system load!", "err", err)
	}
	lb.Add("load01", strconv.FormatFloat(load.Load1, 'f', 1, 64))
	lb.Add("load05", strconv.FormatFloat(load.Load5, 'f', 1, 64))
	lb.Add("load15", strconv.FormatFloat(load.Load15, 'f', 1, 64))

	vmem, err := mem.VirtualMemory()
	if err != nil {
		log.Error("Error getting system memory info!", "err", err)
	}
	lb.Add("mem_perc", strconv.FormatFloat(vmem.UsedPercent, 'f', 1, 64))
	lb.Add("mem_used", strconv.FormatUint(uint64(vmem.Used), 10))
	lb.Add("mem_free", strconv.FormatUint(uint64(vmem.Free), 10))

	return lb.Encode()
}

func main() {
	logPath := os.Getenv("LOGFILE")
	if logPath == "" {
		logPath = "/var/log/system-reporting.json"
	}

	logFile, err := os.OpenFile(
		logPath,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0o644,
	)
	if err != nil {
		log.Fatal("Could't open log file!", "err", err)
	}
	defer logFile.Close()

	if ih := os.Getenv("INFLUX_HOST"); ih != "" {
		config.InfluxHost = ih
	}
	if config.InfluxHost == "" {
		log.Fatal("InfluxHost is unset!")
	}

	mwriter := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mwriter)
	log.SetFormatter(log.LogfmtFormatter)
	log.SetLevel(log.InfoLevel)

	if hostname = os.Getenv("DEVICE"); hostname == "" {
		hostname, err = os.Hostname()
		if err != nil {
			log.Fatal("Error getting device hostname", "err", err)
		}
	}


	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	reqs := make(chan *http.Request)

	ticker := time.NewTicker(5 * time.Second)
	go func() { for {
		select {
		case <-ctx.Done():
			cancel()
			return
		case <-ticker.C:
			log.Info("Recording measurement")
			data := recordMeasurement()
			req, err := http.NewRequest(
				"POST",
				config.InfluxHost,
				bytes.NewBuffer([]byte(data)),
			)
			if err != nil {
				log.Error("Error creating request", "err" ,err)
				continue
			}
			reqs <- req

		}
	}}()

	go func() { for {
		select {
		case <-ctx.Done():
			return
		case req := <-reqs:
			log.Debug("Making request")
			_, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Error("Error sending data", "err", err)
			}
		}
	}}()

	defer ticker.Stop()

	<-ctx.Done()
	log.Info("Stopping...")
}
