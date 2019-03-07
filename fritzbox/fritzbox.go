package fritzbox

import (
    "net/http"
    "crypto/tls"
    "net/url"

    "github.com/coolduke/prometheus-fritzbox-exporter/config"
    
    "github.com/op/go-logging"
    "github.com/bpicode/fritzctl/fritz"
    "github.com/bpicode/fritzctl/logger"
)

type FritzBox struct {
  Log *logging.Logger
  Config *config.ConfigFritzBox
  HomeAuto fritz.HomeAuto
  FritzClient fritz.Client
  Internal fritz.Internal
}

func NewFritzBox(log *logging.Logger, conf config.ConfigFritzBox) (*FritzBox, error) {
  fritzboxUrl, err := url.Parse(conf.Url)
  if err != nil {
    return nil, err
  }

  log.Debugf("Trying %s", conf.Url)
  homeAuto := fritz.NewHomeAuto(
    fritz.SkipTLSVerify(),
    fritz.URL(fritzboxUrl),
    fritz.Credentials(conf.Username, conf.Password),
  )

  l := &logger.Level{}
  l.Set("warn")

  err = homeAuto.Login()
  if err != nil {
    return nil, err
  }

  //build fritzctl client with config backed by our yaml file
  configPtr := &conf.FritzctlConfig
  tlsConfig := &tls.Config{InsecureSkipVerify: conf.FritzctlConfig.Pki.SkipTLSVerify}
  transport := &http.Transport{TLSClientConfig: tlsConfig}
  httpClient := &http.Client{Transport: transport}
  fritzClient := &fritz.Client{Config: configPtr, HTTPClient: httpClient}

  err = fritzClient.Login()
  if err != nil {
    log.Error("Unable to login")
    return nil, err
  }

  internal := fritz.NewInternal(fritzClient)
  
  return &FritzBox{Log: log, Config: &conf, HomeAuto: homeAuto, FritzClient: *fritzClient, Internal: internal}, nil
}

func (fb *FritzBox) LogCurrentTemperatures() error {
  devices, err := fb.HomeAuto.List()
  if err != nil {
    return err
  }
  
  for _, device := range devices.Thermostats() {
    fb.Log.Infof("Current temperature for %s: %s°C", device.Name, device.Thermostat.FmtMeasuredTemperature())
  }
  
  return nil
}

func (fb *FritzBox) SetTemperature(thermostat string, value float64) (error) {
  err := fb.HomeAuto.Temp(value, thermostat)
  if err != nil {
    return err
  }
  
  return nil
}
