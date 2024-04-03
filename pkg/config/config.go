/*
Copyright 2022 Adevinta
*/

// Package config parses and validates the configuration
// needed to run vulcan-tracker.
package config

import (
	"io"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/log"

	"github.com/adevinta/vulcan-tracker/pkg/storage/postgresql"
)

// Config represents all the configuration needed to run the project.
type Config struct {
	API  apiConfig          `toml:"api"`
	Log  logConfig          `toml:"log"`
	PSQL postgresql.ConnStr `toml:"postgresql"`
	AWS  AwsConfig          `toml:"aws"`
}

type apiConfig struct {
	MaxSize     int `toml:"max_size"`
	DefaultSize int `toml:"default_size"`
	Port        int `toml:"port"`
}

type logConfig struct {
	Level string `toml:"level"`
}

// AwsConfig stores the AWS configuration.
type AwsConfig struct {
	ServerCredentialsKey string `toml:"server_credentials_key"`
	Endpoint             string `toml:"endpoint"`
	Region               string `toml:"region"`
}

// ParseConfig parses de config file and set default values when it is needed.
func ParseConfig(cfgFilePath string) (*Config, error) {
	cfgFile, err := os.Open(cfgFilePath)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()

	cfgData, err := io.ReadAll(cfgFile)
	if err != nil {
		return nil, err
	}

	var conf Config
	if _, err := toml.Decode(string(cfgData[:]), &conf); err != nil {
		return nil, err
	}

	if envVar := os.Getenv("PORT"); envVar != "" {
		entVarInt, err := strconv.Atoi(envVar)
		if err != nil {
			return nil, err
		}
		conf.API.Port = entVarInt
	}

	if envVar := os.Getenv("VULCANTRACKER_DB_HOST"); envVar != "" {
		conf.PSQL.Host = envVar
	}

	if envVar := os.Getenv("VULCANTRACKER_DB_PORT"); envVar != "" {
		conf.PSQL.Port = envVar
	}

	if envVar := os.Getenv("VULCANTRACKER_DB_USER"); envVar != "" {
		conf.PSQL.User = envVar
	}

	if envVar := os.Getenv("VULCANTRACKER_DB_NAME"); envVar != "" {
		conf.PSQL.DB = envVar
	}

	if envVar := os.Getenv("VULCANTRACKER_AWS_REGION"); envVar != "" {
		conf.AWS.Region = envVar
	}

	return &conf, nil
}

// ParseLogLvl parses the level of a log from a string to a log.Lvl object.
func ParseLogLvl(lvl string) log.Lvl {
	switch lvl {
	case "ERROR":
		return log.ERROR
	case "WARN":
		return log.WARN
	case "DEBUG":
		return log.DEBUG
	default:
		return log.INFO
	}
}
