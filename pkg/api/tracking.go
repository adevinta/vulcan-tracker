/*
Copyright 2002 Adevinta
*/

package api

import (
	"net/http"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
	"github.com/labstack/echo/v4"
)

func response(c echo.Context, httpStatus int, data interface{}, dataType string, p ...tracking.Pagination) error {
	if data == nil {
		return c.NoContent(http.StatusNoContent)
	}

	resp := map[string]interface{}{}

	// We check if the variadic argument is present.
	if len(p) > 0 {
		// We only use the first element, as we expect only one.
		more := p[0].Total > p[0].Offset+p[0].Limit

		pagination := Pagination{
			Limit:  p[0].Limit,
			Offset: p[0].Offset,
			Total:  p[0].Total,
			More:   more,
		}

		if p[0].Offset > p[0].Total {
			return echo.NewHTTPError(http.StatusNotFound, ErrPageNotFound.Error())
		}

		resp["pagination"] = pagination
	}

	resp[dataType] = data

	return c.JSON(httpStatus, resp)
}

// GetTicket returns a JSON containing a specific ticket.
func (api *API) GetTicket(c echo.Context) error {
	id := c.Param("id")
	teamId := c.QueryParam("teamId")

	// Get the server and the configuration for the teamId.
	configuration, err := api.storage.GetTrackerConfiguration(teamId)
	if err != nil {
		return err
	}

	serverName := configuration.ServerName

	ticket, err := api.trackingServers[serverName].GetTicket(id)
	if err != nil {
		return err
	}

	if ticket.ID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}

// CreateTicket creates a ticket and returns a JSON containing the new ticket.
func (api *API) CreateTicket(c echo.Context) error {
	ticket := new(model.Ticket)
	teamId := c.QueryParam("teamId")

	if err := c.Bind(ticket); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Get the server and the configuration for the teamId.
	configuration, err := api.storage.GetTrackerConfiguration(teamId)
	if err != nil {
		return err
	}

	// Retrieve the necesary values to create a ticket.
	ticket.Project = configuration.Project
	ticket.TicketType = configuration.VulnerabilityIssueType
	serverName := configuration.ServerName

	ticket, err = api.trackingServers[serverName].CreateTicket(ticket)
	if err != nil {
		return err
	}
	if ticket.ID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}

// FixTicket updates a ticket until a "done" state and returns a JSON containing the new ticket.
func (api *API) FixTicket(c echo.Context) error {
	id := c.Param("id")
	teamId := c.QueryParam("teamId")

	// Get the server and the configuration for the teamId,
	configuration, err := api.storage.GetTrackerConfiguration(teamId)
	if err != nil {
		return err
	}

	serverName := configuration.ServerName

	ticket, err := api.trackingServers[serverName].FixTicket(id, configuration.FixedWorkflow)
	if err != nil {
		return err
	}
	if ticket.ID == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}
