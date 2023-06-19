/*
Copyright 2022 Adevinta
*/
package jira

import (
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/labstack/echo/v4"
)

// TrackerClient represents a Jira ticket tracker client.
type TrackerClient struct {
	Client TicketTrackingClient
	Logger echo.Logger
	URL    string
}

// TicketTrackingClient defines the API of the adapter for a third-party client.
type TicketTrackingClient interface {
	GetTicket(id string) (model.Ticket, error)
	FindTicket(projectKey, vulnerabilityIssueType, text string) (model.Ticket, error)
	CreateTicket(ticket model.Ticket) (model.Ticket, error)
}

// New instantiates a new Jira connection.
func New(url, token string, logger echo.Logger) (*TrackerClient, error) {
	jiraClient, err := NewClient(url, token)
	if err != nil {
		return nil, err
	}

	return &TrackerClient{
		Client: jiraClient,
		Logger: logger,
		URL:    url,
	}, nil
}
