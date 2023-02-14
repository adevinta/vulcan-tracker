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
		URL    string
	}
)

// TicketTrackingClient defines the API of the adapter for a third-party client.
type TicketTrackingClient interface {
	GetTicket(id string) (*model.Ticket, error)
	FindTicket(projectKey, vulnerabilityIssueType, text string) (*model.Ticket, error)
	CreateTicket(ticket *model.Ticket) (*model.Ticket, error)
	GetTicketTransitions(id string) ([]model.Transition, error)
	DoTransition(id, idTransition string) error
	DoTransitionWithResolution(id, idTransition, resolution string) error
}

// New instantiates a new Jira connection.
func New(url, user, pass string, logger echo.Logger) (*TC, error) {
	jiraClient, err := NewClient(url, user, pass)
	if err != nil {
		return nil, err
	}

	return &TC{
		Client: jiraClient,
		Logger: logger,
		URL:    url,
	}, nil
}
