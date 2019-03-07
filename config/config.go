package config

import (
    "io/ioutil"

    "gopkg.in/yaml.v2"
    "github.com/op/go-logging"

    fritzctlConfig "github.com/bpicode/fritzctl/config"
)

type Config struct {
  FritzBox         *ConfigFritzBox         `yaml:"fritzbox"`
  Exporter         *ConfigExporter         `yaml:"exporter"`
}

type ConfigFritzBox struct {
  Url      string `yaml:"url"`
  Username string `yaml:"username"`
  Password string `yaml:"password"`
  FritzctlConfig   fritzctlConfig.Config
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

  //provide configuration in a format fritzctl expects it
  //TODO: rewrite URL
  config.FritzBox.FritzctlConfig = fritzctlConfig.Config{
    Net: &fritzctlConfig.Net{Protocol: "https", Host: "fritz.box", Port: "49232"},
    Login: &fritzctlConfig.Login{LoginURL: "/login_sid.lua", Username: config.FritzBox.Username, Password: config.FritzBox.Password},
    Pki: &fritzctlConfig.Pki{SkipTLSVerify: true, CertificateFile: ""},
  }

  return config, nil
}
