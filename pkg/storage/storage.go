/*
Copyright 2022 Adevinta
*/
package storage

import "github.com/adevinta/vulcan-tracker/pkg/model"

type TicketServerStorage interface {
	ServersConf() ([]model.TrackerConfig, error)
	ProjectsConfig() ([]model.ProjectConfig, error)
	ProjectConfig(name string) (*model.ProjectConfig, error)
}

type Storage interface {
	CreateFindingTicket(t model.Ticket) (model.FindingTicket, error)
	GetFindingTicket(findingID, teamID string) (model.FindingTicket, error)
}
