package main

import (
    "fmt"
    "os"
//    "flag"
    "time"
    "strconv"
    "net/http"

    
    "github.com/coolduke/prometheus-fritzbox-exporter/config"
    "github.com/coolduke/prometheus-fritzbox-exporter/fritzbox"

    "github.com/op/go-logging"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var log = logging.MustGetLogger("fritzbox-exporter")
var format = logging.MustStringFormatter(
    `%{color}%{time:15:04:05.000} %{level:-8s} ▶ %{shortpkg:-10s} ▶%{color:reset} %{message}`,
)

var (
  fb *fritzbox.FritzBox
  conf config.Config
)

// Metric name parts.
const (
  namespace = "fritzbox"
  exporter  = "exporter"
)

type Exporter struct {
  duration, error, up           prometheus.Gauge
  totalScrapes                  prometheus.Counter
  scrapeErrors                  *prometheus.CounterVec

  boxinfoLifetime               *prometheus.GaugeVec
  boxinfoReboots                *prometheus.GaugeVec

  internetStats                 *prometheus.GaugeVec

  homeAutoDevicePresent         *prometheus.GaugeVec
  homeAutoDeviceTemperature     *prometheus.GaugeVec
  homeAutoDeviceSwitchState     *prometheus.GaugeVec
  homeAutoDevicePower           *prometheus.GaugeVec
  homeAutoDeviceEnergy          *prometheus.GaugeVec
}

func NewExporter() *Exporter {
  return &Exporter{
    duration: prometheus.NewGauge(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "last_scrape_duration_seconds",
      Help:      "Duration of the last scrape.",
    }),
    totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "scrapes_total",
      Help:      "Total number of scrapes.",
    }),
    scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "scrape_errors_total",
      Help:      "Total number of times an error occured while scraping.",
    }, []string{"collector"}),
    error: prometheus.NewGauge(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "last_scrape_error",
      Help:      "Whether the last scrape resulted in an error.",
    }),
    up: prometheus.NewGauge(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: exporter,
      Name:      "up",
      Help:      "Whether the connection to the FritzBox is established.",
    }),

    boxinfoLifetime: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "boxinfo",
      Name:      "lifetime",
      Help:      "Days since date manufacture.",
    }, []string{"model","firmware_version"}),
    boxinfoReboots: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "boxinfo",
      Name:      "reboots",
      Help:      "Number of reboots since date of manufacture.",
    }, []string{"model","firmware_version"}),


    internetStats: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "internet",
      Name:      "bandwidth",
      Help:      "Bandwidth for Internet access.",
    }, []string{"direction","type"}),

    homeAutoDevicePresent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "homeauto",
      Name:      "device_present",
      Help:      "Whether the device is connected to the FritBox.",
    }, []string{"uuid","name","productname"}),
    homeAutoDeviceTemperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "homeauto",
      Name:      "device_temperature",
      Help:      "Gauge metric with temperature (in Celsius) of connected devices.",
    }, []string{"uuid","name","productname"}),
    homeAutoDeviceSwitchState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "homeauto",
      Name:      "device_switch_state",
      Help:      "Whether the switch device is on or off.",
    }, []string{"uuid","name","productname"}),
    homeAutoDevicePower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "homeauto",
      Name:      "device_power",
      Help:      "Gauge metric with current power (in Watt) of connected devices.",
    }, []string{"uuid","name","productname"}),
    homeAutoDeviceEnergy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
      Namespace: namespace,
      Subsystem: "homeauto",
      Name:      "device_energy",
      Help:      "Counter metric with absolute energy consumption (in Watt hours) since the device started operating.",
    }, []string{"uuid","name","productname"}),
  }
}

func (e *Exporter) ScrapeBoxinfo() error {
  boxinfo, err := fb.Internal.BoxInfo()
  if err != nil {
    return err
  }

  versionString := fmt.Sprintf("%s.%s.%s", boxinfo.FirmwareVersion.OsVersionMajor, boxinfo.FirmwareVersion.OsVersionMinor, boxinfo.FirmwareVersion.OsVersionRevision)

  e.boxinfoLifetime.WithLabelValues(boxinfo.Model.Name, versionString).Set(
    float64(boxinfo.Runtime.Years) * 365.24220 + float64(boxinfo.Runtime.Months) * 30.43685 + float64(boxinfo.Runtime.Days))

  e.boxinfoReboots.WithLabelValues(boxinfo.Model.Name, versionString).Set(float64(boxinfo.Runtime.Reboots))

  return nil
}

func (e *Exporter) ScrapeInternetStats() error {
/*
  stats, err := fb.Internal.InternetStats()
  if err != nil {
    return err
  }

  e.internetStats.WithLabelValues("downstream", "internet").Set(avg(stats.DownstreamInternet))
  e.internetStats.WithLabelValues("downstream", "media").Set(avg(stats.DownStreamMedia))
  e.internetStats.WithLabelValues("upstream", "realtime").Set(avg(stats.UpstreamRealtime))
  e.internetStats.WithLabelValues("upstream", "high").Set(avg(stats.UpstreamHighPriority))
  e.internetStats.WithLabelValues("upstream", "default").Set(avg(stats.UpstreamDefaultPriority))
  e.internetStats.WithLabelValues("upstream", "low").Set(avg(stats.UpstreamLowPriority))
*/
  return nil
}

func (e *Exporter) ScrapeHomeAutoDevices() error {
  devices, err := fb.HomeAuto.List()
  if err != nil {
    return err
  }

  for _, device := range devices.Devices {
    //device up?
    e.homeAutoDevicePresent.WithLabelValues(device.Identifier, device.Name, device.Productname).Set(float64(device.Present))

    //temperature if available
    if device.CanMeasureTemp() {
      temperature, err := strconv.ParseFloat(device.Temperature.FmtCelsius(), 64)
      if err == nil {
        e.homeAutoDeviceTemperature.WithLabelValues(device.Identifier, device.Name, device.Productname).Set(temperature)
      }
    }

    //switch state if available
    if device.IsSwitch() {
      switchState, err := strconv.ParseFloat(device.Switch.State, 64)
      if err == nil {
        e.homeAutoDeviceSwitchState.WithLabelValues(device.Identifier, device.Name, device.Productname).Set(switchState)
      }
    }

    //power/energy if available
    if device.CanMeasurePower() {
      power, err := strconv.ParseFloat(device.Powermeter.Power, 64)
      if err == nil {
        e.homeAutoDevicePower.WithLabelValues(device.Identifier, device.Name, device.Productname).Set(power * 0.001)
      }

      energy, err := strconv.ParseFloat(device.Powermeter.Energy, 64)
      if err == nil {
        e.homeAutoDeviceEnergy.WithLabelValues(device.Identifier, device.Name, device.Productname).Set(energy)
      }
    }
  }
  return nil
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

  err = e.Connect()
  if err == nil {
    e.error.Set(0)

    e.ScrapeBoxinfo()
    e.boxinfoLifetime.Collect(ch)
    e.boxinfoReboots.Collect(ch)

    e.ScrapeInternetStats()
    e.internetStats.Collect(ch)

    e.ScrapeHomeAutoDevices()
    e.homeAutoDevicePresent.Collect(ch)
    e.homeAutoDeviceTemperature.Collect(ch)
    e.homeAutoDeviceSwitchState.Collect(ch)
    e.homeAutoDevicePower.Collect(ch)
    e.homeAutoDeviceEnergy.Collect(ch)

    
  } else {
    e.error.Set(1)
  }
  e.up.Collect(ch)


  ch <- e.duration
  ch <- e.totalScrapes
  ch <- e.error
  e.scrapeErrors.Collect(ch)
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
  e.duration.Describe(ch)
  e.totalScrapes.Describe(ch)
  e.scrapeErrors.Describe(ch)
  e.up.Describe(ch)

  e.homeAutoDevicePresent.Describe(ch)
  e.homeAutoDeviceTemperature.Describe(ch)
  e.homeAutoDeviceSwitchState.Describe(ch)
  e.homeAutoDevicePower.Describe(ch)
  e.homeAutoDeviceEnergy.Describe(ch)
}

func (e *Exporter) Connect() error {
  var err error

  log.Debug("Scraping...")

  if fb == nil {
    fb, err = fritzbox.NewFritzBox(log, *conf.FritzBox)
    if err != nil {
      log.Error(err.Error())
      return err
    }
  } else {
    err = fb.HomeAuto.Login()
    if err != nil {
      return err
    }
  }

  return nil
}

func (e *Exporter) Handler(w http.ResponseWriter, r *http.Request) {
  promhttp.Handler().ServeHTTP(w, r)
}

func main() {
    var err error

    logBackend := logging.NewLogBackend(os.Stderr, "", 0)
    logBackendFormatter := logging.NewBackendFormatter(logBackend, format)
    logBackendLeveled := logging.AddModuleLevel(logBackendFormatter)
    logBackendLeveled.SetLevel(logging.DEBUG, "")
    logging.SetBackend(logBackendLeveled)

    //load configuration
    conf, err = config.GetConfig(log, "prometheus-fritzbox-exporter.yml")
    if err != nil {
      log.Error(err.Error())
      os.Exit(1)
    }
    
    exporter := NewExporter()
    prometheus.MustRegister(exporter)

    http.HandleFunc("/metrics", exporter.Handler)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {w.Write([]byte("Use /metrics"))})

    log.Info("Listening on", conf.Exporter.ListenAddress)
    log.Error(http.ListenAndServe(conf.Exporter.ListenAddress, nil))
}

func avg(slice[] float64) float64 {
    sum := 0.0

    for _, value := range slice {
      sum += value
    }

    return sum / float64(len(slice))
}