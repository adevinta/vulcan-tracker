/*
Copyright 2022 Adevinta
*/

package jira

import (
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/labstack/echo/v4"
)

type (
	TC struct {
		Client TicketTrackingClient
		Logger echo.Logger
	}
)

// TicketTrackingClient defines the API of the adapter for a third-party client.
type TicketTrackingClient interface {
	GetTicket(id string) (*model.Ticket, error)
	CreateTicket(ticket *model.Ticket, issueType string) (*model.Ticket, error)
	GetTicketTransitions(id string) ([]model.Transition, error)
	DoTransition(id string, idTransition string) error
}

// New instantiates a new Jira connection.
func New(serverConf model.TrackerConfig, logger echo.Logger) (*TC, error) {

	jiraClient, err := NewClient(serverConf.Url, serverConf.User, serverConf.Pass)
	if err != nil {
		return nil, err
	}

	return &TC{
		Client: jiraClient,
		Logger: logger,
	}, nil
}
