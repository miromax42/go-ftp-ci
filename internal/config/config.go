package config

import (
	"fmt"
	"github.com/miromax42/go-ftp-ci/internal/watcher"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Ftp   watcher.Config      `yaml:"ftp"`
	Tasks map[string][]string `yaml:"tasks"`
}

func New(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return c, nil
}
