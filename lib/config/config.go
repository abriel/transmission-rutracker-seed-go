package config

import(
  "github.com/kirsle/configdir"
  "github.com/mcuadros/go-defaults"
  "github.com/go-playground/validator/v10"
  "gopkg.in/yaml.v3"
  "net/url"
  "path"
  "os"
  "slices"
  "fmt"
)

type Config struct {
  BtUriRaw string `default:"http://127.0.0.1:9091/transmission/rpc" yaml:"bt_uri" validate:"http_url" desc:"Bittorrent connection string"`
  UserAgent string `default:"Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0" yaml:"user_agent" desc:"HTTP User-Agent"`
}

func ConfigFile() string {
  return path.Join(configdir.LocalConfig(), "transmission-rutracker-seed-go.yml")
}

func ConfigFileNext() string {
  return fmt.Sprint(ConfigFile(), ".next")
}

func New(args ...string) (*Config, error) {
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

  if !slices.Contains(args, "skip_validation") {
    validate := validator.New()
    err = validate.Struct(myconfig)
    if err != nil {
      return myconfig, err
    }
  }

  return myconfig, nil
}

func (c Config) Save() error {
  if err := validator.New().Struct(c) ; err != nil {
    return fmt.Errorf("Validation failed for the config %v due to: %v", c, err)
  }

  fd, err := os.Create(ConfigFileNext())
  if err != nil {
    return fmt.Errorf("Error while opening new configuration file %v: %v", ConfigFileNext(), err)
  }

  yaml_encoder := yaml.NewEncoder(fd)
  if err := yaml_encoder.Encode(c) ; err != nil {
    return fmt.Errorf("Error while encoding yaml: %v", err)
  }

  fd.Close()
  if err := os.Rename(ConfigFileNext(), ConfigFile()) ; err != nil {
    return fmt.Errorf("Error while renaming %v to %v: %v", ConfigFileNext(), ConfigFile(), err)
  }

  return nil
}

func (c Config) BtUri() (endpoint *url.URL) {
  endpoint, err := url.Parse(c.BtUriRaw)
  if err != nil {
    // We really don't expect to have that error
    panic(err)
  }
  return
}

func (c Config) BtProtocolisHttp() bool {
  return c.BtUri().Scheme == "http"
}

func (c Config) BtProtocolisHttps() bool {
  return c.BtUri().Scheme == "https"
}

func (c Config) BtUsername() string {
  return c.BtUri().User.Username()
}

func (c Config) BtPassword() string {
  password, _ := c.BtUri().User.Password()
  return password
}

func (c Config) BtHost() string {
  return c.BtUri().Hostname()
}

func (c Config) BtPort() string {
  return c.BtUri().Port()
}

func (c Config) BtPath() string {
  return c.BtUri().Path
}
