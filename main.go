package main

import (
//    "fmt"
    "os"
//    "flag"
    "net/http"

    
    "github.com/coolduke/prometheus-fritzbox-exporter/config"
    "github.com/coolduke/prometheus-fritzbox-exporter/fritzbox"

    "github.com/op/go-logging"
    "github.com/prometheus/client_golang/prometheus"
)

var log = logging.MustGetLogger("fritzbox-exporter")
var format = logging.MustStringFormatter(
    `%{color}%{time:15:04:05.000} %{level:-8s} ▶ %{shortpkg:-10s} ▶%{color:reset} %{message}`,
)

var (
  fb *fritzbox.FritzBox
)

// Metric name parts.
const (
  namespace = "fritzbox"
  exporter  = "exporter"
)

type Exporter struct {
  duration, error prometheus.Gauge
  totalScrapes    prometheus.Counter
  scrapeErrors    *prometheus.CounterVec
  temperature     *prometheus.GaugeVec
}

func NewExporter() *Exporter {
  return &Exporter{
    duration: prometheus.NewGauge(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "last_scrape_duration_seconds",
      Help:      "Duration of the last scrape of metrics from Oracle DB.",
    }),
    totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "scrapes_total",
      Help:      "Total number of times Oracle DB was scraped for metrics.",
    }),
    scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "scrape_errors_total",
      Help:      "Total number of times an error occured scraping a Oracle database.",
    }, []string{"collector"}),
    error: prometheus.NewGauge(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "last_scrape_error",
      Help:      "Whether the last scrape of metrics from Oracle DB resulted in an error (1 for error, 0 for success).",
    }),
    temperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Name:      "temperature",
      Help:      "Gauge metric with temperature (in Celsius) of connected devices.",
    }, []string{"device","type"}),
  }
}

func (e *Exporter) ScrapeTemperature() {
  fb.LogCurrentTemperatures()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
  var err error

  e.totalScrapes.Inc()
  defer func(begun time.Time) {
    e.duration.Set(time.Since(begun).Seconds())
    if err == nil {
      e.error.Set(0)
    } else {
      e.error.Set(1)
    }
  }(time.Now())

  e.Connect()
  e.up.Collect(ch)

  e.ScrapeUptime()
  e.uptime.Collect(ch)

  e.ScrapeTemperature()
  e.temperature.Collect(ch)

  ch <- e.duration
  ch <- e.totalScrapes
  ch <- e.error
  e.scrapeErrors.Collect(ch)
}

func main() {
    var err error

    logBackend := logging.NewLogBackend(os.Stderr, "", 0)
    logBackendFormatter := logging.NewBackendFormatter(logBackend, format)
    logBackendLeveled := logging.AddModuleLevel(logBackendFormatter)
    logBackendLeveled.SetLevel(logging.DEBUG, "")
    logging.SetBackend(logBackendLeveled)

    //load configuration
    config, err := config.GetConfig(log, "prometheus-fritzbox-exporter.yml")
    if err != nil {
      log.Error(err.Error())
      os.Exit(1)
    }
    
    fb, err = fritzbox.NewFritzBox(log, *config.FritzBox)
    if err != nil {
      log.Error(err.Error())
      os.Exit(1)
    }

    exporter := NewExporter()
    prometheus.MustRegister(exporter)

    http.HandleFunc("/metrics", exporter.Handler)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {w.Write("Use /metrics")})

    log.Info("Listening on", config.Exporter.ListenAddress)
    log.Error(http.ListenAndServe(config.Exporter.ListenAddress, nil))
    
//    err = fritzbox.SetTemperature("Wohnzimmer", 17)
//    if err != nil {
//      log.Error(err.Error())
//      os.Exit(1)
//    }
}
