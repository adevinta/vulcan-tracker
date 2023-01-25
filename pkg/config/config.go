/*
Copyright 2022 Adevinta
*/
package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/log"
)

type Server struct {
	Url   string `toml:"url"`
	User  string `toml:"user"`
	Token string `toml:"token"`
	Kind  string `toml:"kind"`
}

type Team struct {
	Server                 string   `toml:"server"`
	Project                string   `toml:"project"`
	VulnerabilityIssueType string   `toml:"vulnerability_issue_type"`
	FixWorkflow            []string `toml:"fix_workflow"`
	WontFixWorkflow        []string `toml:"wontfix_workflow"`
	AutoCreate             bool     `toml:"auto_create"` // create tickets automatically from a stream

}

type KafkaConfig struct {
	RetryDuration time.Duration `toml:"retry_duration"`
	KafkaBroker   string        `toml:"broker"`
	KafkaGroupID  string        `toml:"group_id"`
	KafkaUsername string        `toml:"username"`
	KafkaPassword string        `toml:"password"`
}

type Config struct {
	API     apiConfig         `toml:"api"`
	Servers map[string]Server `toml:"servers"`
	Teams   map[string]Team   `toml:"teams"`
	Kafka   KafkaConfig       `toml:"kafka"`
	Log     logConfig         `toml:"log"`
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
		envVarInt, err := strconv.Atoi(envVar)
		if err != nil {
			return nil, err
		}
		conf.API.Port = envVarInt
	}

	if envVar := os.Getenv("KAFKA_BROKER"); envVar != "" {
		conf.Kafka.KafkaBroker = envVar
	}

	if envVar := os.Getenv("RETRY_DURATION"); envVar != "" {
		conf.Kafka.RetryDuration, err = time.ParseDuration(envVar)
		if err != nil {
			return nil, fmt.Errorf("invalid retry duration: %w", err)
		}
	}

	if envVar := os.Getenv("KAFKA_GROUP_ID"); envVar != "" {
		conf.Kafka.KafkaGroupID = envVar
	}

	if envVar := os.Getenv("KAFKA_USERNAME"); envVar != "" {
		conf.Kafka.KafkaUsername = envVar
	}

	if envVar := os.Getenv("KAFKA_PASSWORD"); envVar != "" {
		conf.Kafka.KafkaPassword = envVar
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
