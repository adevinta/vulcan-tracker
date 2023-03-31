/*
Copyright 2022 Adevinta
*/
package storage

import "github.com/adevinta/vulcan-tracker/pkg/model"

// TicketServerStorage manages the storage of the ticket trackers configuration.
type TicketServerStorage interface {
	FindServerConf(serverID string) (model.TrackerConfig, error)
	FindProjectConfigByTeamID(teamID string) (model.ProjectConfig, error)
}

// Storage manages the storage of the project data.
type Storage interface {
	CreateFindingTicket(t model.Ticket) (model.FindingTicket, error)
	GetFindingTicket(findingID, teamID string) (model.FindingTicket, error)

	Healthcheck() error
}
