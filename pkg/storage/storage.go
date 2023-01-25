/*
Copyright 2022 Adevinta
*/
package storage

import "github.com/adevinta/vulcan-tracker/pkg/model"

type Storage interface {
	ServersConf() ([]model.TrackerConfig, error)
	ProjectsConfig() ([]model.ProjectConfig, error)
	ProjectConfig(name string) (*model.ProjectConfig, error)
}
