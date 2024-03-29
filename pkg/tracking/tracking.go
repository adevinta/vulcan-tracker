/*
Copyright 2022 Adevinta
*/

// Package tracking manages a ticket tracker server.
package tracking

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/secrets"
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking/jira"
)

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

	ttClient, err = jira.New(serverConf.URL, serverConf.Token, logger)
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
	if errors.Is(err, sql.ErrNoRows) {
		return model.TrackerConfig{}, &vterrors.TrackingError{
			HTTPStatusCode: http.StatusNotFound,
			Err:            fmt.Errorf("server not found: %w", err),
		}
	}
	if err != nil {
		return model.TrackerConfig{}, err
	}
	credentials, err := ts.secretsService.GetServerCredentials(serverConfig.ID)
	if err != nil {
		return model.TrackerConfig{}, err
	}
	serverConfig.Token = credentials.Token

	return serverConfig, nil
}

// ProjectConfigByTeamID retrieves all the needed configuration to access a ticket tracker project for a
// specific team.
func (ts *TS) ProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {
	// Get the server and the configuration for the teamId.
	configuration, err := ts.ticketServerStorage.FindProjectConfigByTeamID(teamID)
	if errors.Is(err, sql.ErrNoRows) {
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
