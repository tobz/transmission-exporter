package main

import (
	"net/http"

	arg "github.com/alexflint/go-arg"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	transmission "github.com/tobz/transmission-exporter"

	"go.uber.org/zap"
)

// Config gets its content from env and passes it on to different packages
type Config struct {
	TransmissionAddr     string `arg:"-h,--transmission-addr,env:TRANSMISSION_ADDR" default:"http://localhost:9091/transmission"`
	TransmissionUsername string `arg:"-P,--transmission-username,env:TRANSMISSION_USERNAME"`
	TransmissionPassword string `arg:"-u,--transmission-password,env:TRANSMISSION_PASSWORD"`
	MetricsListenAddr    string `arg:"-l,env:METRICS_LISTEN_ADDR" default:":19091"`
	MetricsPath          string `arg:"-p,env:METRICS_PATH" default:"/metrics"`
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting transmission-exporter.")

	// Try loading any env files first, which populates $ENV before loading the configuration.
	err := godotenv.Load()
	if err != nil {
		logger.Debug("No .env file was present.")
	}

	// Now load our configuration, either via environment variables or CLI flags.
	conf := Config{}
	if err = arg.Parse(&conf); err != nil {
		logger.Fatal("Failed to parse command-line arguments.", zap.Error(err))
	}

	// Configure and construct our Transmission client.
	var user *transmission.User
	if conf.TransmissionUsername != "" && conf.TransmissionPassword != "" {
		user = &transmission.User{
			Username: conf.TransmissionUsername,
			Password: conf.TransmissionPassword,
		}
	}

	client, err := transmission.New(logger, conf.TransmissionAddr, user)
	if err != nil {
		logger.Error("Failed to construct Transmission client.", zap.Error(err))

	}

	// Wire up the Prometheus SDK to our various collectors, and serve the metrics endpoint over HTTP.
	prometheus.MustRegister(NewTorrentCollector(logger, client))
	prometheus.MustRegister(NewSessionCollector(logger, client))
	prometheus.MustRegister(NewSessionStatsCollector(logger, client))

	http.Handle(conf.MetricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Transmission Exporter</title></head>
			<body>
			<h1>Transmission Exporter</h1>
			<p><a href="` + conf.MetricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	err = http.ListenAndServe(conf.MetricsListenAddr, nil)
	if err != nil {
		logger.Fatal("Failed to serve metrics endpoint.", zap.Error(err))
	}
}

func NumericBool(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
