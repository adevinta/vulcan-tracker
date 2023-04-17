/*
Copyright 2022 Adevinta
*/
package tracking

import (
	"database/sql"
	"fmt"
	"net/http"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/labstack/echo/v4"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/secrets"
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking/jira"
)

// Filter holds query filtering information.
type Filter struct {
	// TODO: Not specified yet
}

// SortBy holds information for the
// query sorting criteria.
type SortBy struct {
	Field string
	Order string
}

// Pagination holds database pagination information.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// TicketTracker defines the interface for high level querying data from ticket tracker.
type TicketTracker interface {
	GetTicket(id string) (model.Ticket, error)
	FindTicketByFindingAndTeam(projectKey, vulnerabilityIssueType, findingID string, teamID string) (model.Ticket, error)
	CreateTicket(ticket model.Ticket) (model.Ticket, error)
}

// TicketTrackerBuilder builds clients to access ticket trackers.
type TicketTrackerBuilder interface {
	GenerateTicketTrackerClient(ticketServer TicketServer, teamID string, logger echo.Logger) (TicketTracker, error)
}

// TTBuilder represents a builder of clients to access ticket trackers.
type TTBuilder struct {
}

// GenerateTicketTrackerClient generates a ticket tracker client.
func (ttb *TTBuilder) GenerateTicketTrackerClient(ticketServer TicketServer, teamID string, logger echo.Logger) (TicketTracker, error) {
	projectConfig, err := ticketServer.ProjectConfigByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	var serverConf model.TrackerConfig
	serverConf, err = ticketServer.ServerConf(projectConfig.TrackerConfigID)
	if err != nil {
		return nil, err
	}

	var ttClient TicketTracker

	ttClient, err = jira.New(serverConf.URL, serverConf.User, serverConf.Pass, logger)
	if err != nil {
		return nil, err
	}
	return ttClient, nil
}

// TicketServer manages the access to a ticket tracker server.
type TicketServer interface {
	ServerConf(serverID string) (model.TrackerConfig, error)
	ProjectConfigByTeamID(teamID string) (model.ProjectConfig, error)
}

// TS represents a service that manages the access to a ticket tracker server.
type TS struct {
	ticketServerStorage storage.TicketServerStorage
	secretsService      secrets.Secrets
	Logger              echo.Logger
}

// New creates a new instance to red the configuration from a toml file.
func New(ticketServerStorage storage.TicketServerStorage, secretsService secrets.Secrets, logger echo.Logger) (*TS, error) {
	return &TS{
		ticketServerStorage: ticketServerStorage,
		secretsService:      secretsService,
		Logger:              logger,
	}, nil
}

// ServerConf retrieves all the needed configuration to access a ticket tracker server.
func (ts *TS) ServerConf(serverID string) (model.TrackerConfig, error) {
	serverConfig, err := ts.ticketServerStorage.FindServerConf(serverID)
	if err == sql.ErrNoRows {
		return model.TrackerConfig{}, &vterrors.TrackingError{
			HTTPStatusCode: http.StatusNotFound,
			Err:            fmt.Errorf("project not found: %w", err),
		}
	}
	if err != nil {
		return model.TrackerConfig{}, err
	}
	credentials, err := ts.secretsService.GetServerCredentials(serverConfig.ID)
	if err != nil {
		return model.TrackerConfig{}, err
	}
	serverConfig.User = credentials.User
	serverConfig.Pass = credentials.Password

	return serverConfig, nil
}

// ProjectConfigByTeamID retrieves all the needed configuration to access a ticket tracker project for a
// specific team.
func (ts *TS) ProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {
	// Get the server and the configuration for the teamId.
	configuration, err := ts.ticketServerStorage.FindProjectConfigByTeamID(teamID)
	if err == sql.ErrNoRows {
		return model.ProjectConfig{}, &vterrors.TrackingError{
			HTTPStatusCode: http.StatusNotFound,
			Err:            fmt.Errorf("project not found: %w", err),
		}
	}
	if err != nil {
		return model.ProjectConfig{}, err
	}
	return configuration, nil
}
