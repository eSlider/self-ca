package config

import (
	_ "embed"
	"os"

	"github.com/eslider/go-config/env"
	"github.com/eslider/go-config/yaml"

	"github.com/eSlider/self-ca/internal/ca"
)

//go:embed config.yml
var defaultConfigYAML []byte

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Data   DataConfig   `mapstructure:"data"`
	Setup  SetupConfig  `mapstructure:"setup"`
}

type ServerConfig struct {
	APIAddr  string `mapstructure:"apiaddr"`
	TLSAddr  string `mapstructure:"tlsaddr"`
	TLSCert  string `mapstructure:"tlscert"`
	TLSKey   string `mapstructure:"tlskey"`
}

type DataConfig struct {
	Dir string `mapstructure:"dir"`
}

type SetupConfig struct {
	CA     SubjectConfig  `mapstructure:"ca"`
	Server LeafSubject    `mapstructure:"server"`
	Output OutputPaths    `mapstructure:"output"`
}

type SubjectConfig struct {
	CommonName   string `mapstructure:"commonname"`
	Organization string `mapstructure:"organization"`
	Country      string `mapstructure:"country"`
	Province     string `mapstructure:"province"`
	Locality     string `mapstructure:"locality"`
	ValidYears   int    `mapstructure:"validyears"`
}

type LeafSubject struct {
	CommonName  string   `mapstructure:"commonname"`
	DNSNames    []string `mapstructure:"dnsnames"`
	IPAddresses []string `mapstructure:"ipaddresses"`
	ValidYears  int      `mapstructure:"validyears"`
}

type OutputPaths struct {
	CACert     string `mapstructure:"cacert"`
	ServerCert string `mapstructure:"servercert"`
	ServerKey  string `mapstructure:"serverkey"`
}

func Default() Config {
	var cfg Config
	if err := yaml.New(yaml.WithBytes(defaultConfigYAML)).Unmarshal(&cfg); err != nil {
		panic("config: invalid embedded defaults: " + err.Error())
	}
	return cfg
}

func Load(path string) (Config, error) {
	cfg := Default()

	if _, err := os.Stat(path); err == nil {
		y := yaml.New(yaml.WithFile(path))
		if err := y.Unmarshal(&cfg); err != nil {
			return Config{}, err
		}
	} else if !os.IsNotExist(err) {
		return Config{}, err
	}

	e := env.New(env.WithCurrentEnvironment())
	if err := e.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (s SetupConfig) CAOptions() ca.CAOptions {
	return ca.CAOptions{
		CommonName:   s.CA.CommonName,
		Organization: s.CA.Organization,
		Country:      s.CA.Country,
		Province:     s.CA.Province,
		Locality:     s.CA.Locality,
		ValidYears:   s.CA.ValidYears,
	}
}

func (s SetupConfig) LeafOptions() ca.LeafOptions {
	return ca.LeafOptions{
		CommonName:  s.Server.CommonName,
		DNSNames:    s.Server.DNSNames,
		IPAddresses: s.Server.IPAddresses,
		ValidYears:  s.Server.ValidYears,
	}
}
