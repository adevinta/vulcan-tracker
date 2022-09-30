/*
Copyright 2022 Adevinta
*/

package tracking

import (
	"strings"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/tracking/jira"
	"github.com/labstack/echo/v4"
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

// TicketTracking defines the interface for high level querying data from ticket tracker.
type TicketTracking interface {
	GetTicket(id string) (*model.Ticket, error)
	CreateTicket(ticket *model.Ticket) (*model.Ticket, error)
	GetTransitions(id string) ([]model.Transition, error)
	FixTicket(id string, workflow []string) (*model.Ticket, error)
}

const jiraKind = "jira"

// GenerateServerClients instanciates a client for every server passed as argument.
func GenerateServerClients(serverConfs []model.TrackerServerConf, logger echo.Logger) (map[string]TicketTracking, error) {

	clients := make(map[string]TicketTracking)
	for _, server := range serverConfs {
		var client TicketTracking
		var err error

		switch kind := strings.ToLower(server.Kind); kind {
		case jiraKind:
			client, err = jira.New(server, logger)
		}
		// TODO: More kind of trackers coming in the future
		if err != nil {
			return nil, err
		}

		clients[server.Name] = client

	}
	return clients, nil
}
