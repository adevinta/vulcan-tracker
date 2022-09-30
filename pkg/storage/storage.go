package storage

import "github.com/adevinta/vulcan-tracker/pkg/model"

type Storage interface {
	ListTrackerServersConf() ([]model.TrackerServerConf, error)
	ListTrackerConfigurations() ([]model.TrackerConfiguration, error)
	GetTrackerConfiguration(name string) (*model.TrackerConfiguration, error)
}
