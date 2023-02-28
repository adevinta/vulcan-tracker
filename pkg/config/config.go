/*
Copyright 2022 Adevinta
*/

package config

import (
	"io"
	"os"
	"strconv"

	"github.com/adevinta/vulcan-tracker/pkg/storage/postgresql"

	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/log"
)

type Server struct {
	Name  string `toml:"name"`
	Url   string `toml:"url"`
	User  string `toml:"user"`
	Token string `toml:"token"`
	Kind  string `toml:"kind"`
}

type Project struct {
	Name                   string   `toml:"name"`
	ServerID               string   `toml:"server_id"`
	TeamID                 string   `toml:"team_id"`
	Project                string   `toml:"project"`
	VulnerabilityIssueType string   `toml:"vulnerability_issue_type"`
	FixWorkflow            []string `toml:"fix_workflow"`
	WontFixWorkflow        []string `toml:"wontfix_workflow"`
}

type Config struct {
	API      apiConfig          `toml:"api"`
	Servers  map[string]Server  `toml:"servers"`
	Projects map[string]Project `toml:"projects"`
	Log      logConfig          `toml:"log"`
	PSQL     postgresql.ConnStr `toml:"postgresql"`
}

type apiConfig struct {
	MaxSize     int `toml:"max_size"`
	DefaultSize int `toml:"default_size"`
	Port        int `toml:"port"`
}

type logConfig struct {
	Level string `toml:"level"`
}

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
	return &conf, nil
}

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
