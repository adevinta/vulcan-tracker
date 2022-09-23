/*
Copyright 2022 Adevinta
*/

package main

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/adevinta/vulcan-jira-api/pkg/issues/jira"
	"github.com/labstack/gommon/log"
)

type config struct {
	API  apiConfig    `toml:"api"`
	Jira jira.ConnStr `toml:"jira"`
	Log  logConfig    `toml:"log"`
}

type apiConfig struct {
	MaxSize     int `toml:"max_size"`
	DefaultSize int `toml:"default_size"`
}

type logConfig struct {
	Level string `toml:"level"`
}

func parseConfig(cfgFilePath string) (*config, error) {
	cfgFile, err := os.Open(cfgFilePath)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()

	cfgData, err := io.ReadAll(cfgFile)
	if err != nil {
		return nil, err
	}

	var conf config
	if _, err := toml.Decode(string(cfgData[:]), &conf); err != nil {
		return nil, err
	}

	if envVar := os.Getenv("JIRA_URL"); envVar != "" {
		conf.Jira.Url = envVar
	}

	if envVar := os.Getenv("JIRA_USER"); envVar != "" {
		conf.Jira.User = envVar
	}

	if envVar := os.Getenv("JIRA_TOKEN"); envVar != "" {
		conf.Jira.Token = envVar
	}

	if envVar := os.Getenv("JIRA_VULNERABILITY_ISSUE_TYPE"); envVar != "" {
		conf.Jira.VulnerabilityIssueType = envVar
	}

	if envVar := os.Getenv("JIRA_PROJECT"); envVar != "" {
		conf.Jira.Project = envVar
	}

	return &conf, nil
}

func parseLogLvl(lvl string) log.Lvl {
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
