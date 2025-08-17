package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gosnmp/gosnmp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Target struct {
	IP   string `mapstructure:"ip"`
	Room string `mapstructure:"room"`
}

type config struct {
	Targets   []Target
	Community string
}

type Collector struct {
	Ip        string
	Community string
	Room      string
	Logger    *zap.Logger
}

// Collect implements prometheus.Collector.
func (c Collector) Collect(metrics chan<- prometheus.Metric) {
	snmp := gosnmp.GoSNMP{}
	snmp.Context = context.Background()
	snmp.Community = c.Community
	snmp.Version = gosnmp.Version1
	snmp.Target = c.Ip
	snmp.Port = 161
	snmp.Transport = "udp"
	snmp.Timeout = 30 * time.Second
	snmp.MaxRepetitions = 50
	snmp.Retries = 3
	snmp.OnRetry = func(s *gosnmp.GoSNMP) {
		c.Logger.Warn("SNMP retry", zap.String("ip", c.Ip))
	}
	err := snmp.Connect()
	if err != nil {
		c.Logger.Error("Error connecting to SNMP target", zap.String("ip", c.Ip), zap.Error(err))
		return
	}
	defer snmp.Conn.Close()

	data, err := snmp.WalkAll("1.3.6.1.4.1.5040.1.2.6.1.3.1.1")
	if err != nil {
		c.Logger.Error("Error walking SNMP data", zap.String("ip", c.Ip), zap.Error(err))
		return
	}

	for x, p := range data {
		data := ""
		switch p.Value.(type) {
		case string:
			data = p.Value.(string)
		case []uint8:
			data = string(p.Value.([]uint8))
		}
		if strings.Contains(data, "--") {
			continue
		}

		data = strings.TrimSpace(strings.ReplaceAll(data, ",", "."))

		floatValue, err := strconv.ParseFloat(data, 32)
		if err != nil {
			continue
		}

		metric := prometheus.MustNewConstMetric(prometheus.NewDesc(
			"wut_temperature",
			"Temperature reading from WUT sensor",
			[]string{"room", "sensor"},
			nil,
		), prometheus.GaugeValue,
			floatValue,
			strings.ToLower(c.Room), strconv.Itoa(x+1),
		)

		metrics <- metric
	}

}

// Describe implements prometheus.Collector.
func (c Collector) Describe(descs chan<- *prometheus.Desc) {
	descs <- prometheus.NewDesc("wut_temperature", "", []string{"room", "sensor"}, prometheus.Labels{})
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/wut-temperature-exporter/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Panic("No valid configuration found", zap.Error(err))
	}

	config := config{}

	err = viper.Unmarshal(&config)
	if err != nil {
		logger.Panic("No valid configuration found", zap.Error(err))
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		target := query.Get("target")
		if len(query["target"]) != 1 || target == "" {
			http.Error(w, "'target' parameter must be specified once", http.StatusBadRequest)
			return
		}

		registry := prometheus.NewRegistry()
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

		room := ""
		ip := ""
		for _, x := range config.Targets {
			if strings.EqualFold(x.Room, target) || x.IP == target {
				room = x.Room
				ip = x.IP
				break
			}
		}

		if ip == "" {
			logger.Error("No target found", zap.String("target", target))
			http.Error(w, "Not found", 404)
			return
		}

		c := Collector{Ip: ip, Room: room, Community: config.Community, Logger: logger}
		registry.MustRegister(c)
		h.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: ":9191", Handler: nil}
	go func() {
		listenErr := server.ListenAndServe()
		if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			logger.Error("Error starting server", zap.Error(listenErr))
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-interrupt
	logger.Sugar().Infof("Shutting down server. Got signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", zap.Error(err))
	}
	logger.Info("Server stopped")
}
