package config

import(
  "github.com/kirsle/configdir"
  "github.com/mcuadros/go-defaults"
  "github.com/go-playground/validator/v10"
  "gopkg.in/yaml.v3"
  "net/url"
  "path"
  "os"
)

type Config struct {
  BtUriRaw string `default:"http://127.0.0.1:9091/transmission/rpc" yaml:"bt_uri" validate:"http_url"`
  UserAgent string `default:"Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0" yaml:"user_agent"`
}

func ConfigFile() string {
  return path.Join(configdir.LocalConfig(), "transmission-rutracker-seed-go.yml")
}

func New() (*Config, error) {
  myconfig := new(Config)
  defaults.SetDefaults(myconfig)

  fd, err := os.Open(ConfigFile())
  defer fd.Close()

  if err == nil {
    decoder := yaml.NewDecoder(fd)
    err = decoder.Decode(myconfig)
    if err != nil {
      return myconfig, err
    }
  }

  validate := validator.New()
  err = validate.Struct(myconfig)
  if err != nil {
    return myconfig, err
  }

  return myconfig, nil
}

func (c *Config) BtUri() (endpoint *url.URL) {
  endpoint, err := url.Parse(c.BtUriRaw)
  if err != nil {
    // We really don't expect to have that error
    panic(err)
  }
  return
}
