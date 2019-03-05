package config

import (
    "io/ioutil"

    "gopkg.in/yaml.v2"
    "github.com/op/go-logging"
)

type Config struct {
  FritzBox         *ConfigFritzBox         `yaml:"fritzbox"`
  Exporter         *ConfigExporter         `yaml:"exporter"`
}

type ConfigFritzBox struct {
  Url      string `yaml:"url"`
  Username string `yaml:"username"`
  Password string `yaml:"password"`
}

type ConfigExporter struct {
  ListenAddress string `yaml:"listenAddress"`
}

func GetConfig(log *logging.Logger, filename string) (Config, error) {
  log.Noticef("Reading configuration from: %s", filename)

  bytes, err := ioutil.ReadFile(filename)
  if err != nil {
    return Config{}, err
  }

  var config Config

  err = yaml.Unmarshal(bytes, &config)
  if err != nil {
    return Config{}, err
  }

  //TODO: validate configuration

  return config, nil
}
