package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const defaultFileLocation = "prom.yaml"

type Config struct {
	Unchanged Target  `yaml:"unchanged"`
	Changed   Target  `yaml:"changed"`
	Queries   []Query `yaml:"queries"`
}

type Target struct {
	Name  string `yaml:"name"`
	Host  string `yaml:"host"`
	Port  string `yaml:"port"`
	Token string `yaml:"token"`
}

type Query struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

func NewConfig(file string) (Config, error) {
	if file == "" {
		file = defaultFileLocation
	}
	config := Config{}
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error:", err)
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("Error:", err)
		return config, err
	}
	return config, nil
}
