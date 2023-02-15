/*
Copyright 2023 Adevinta
*/
package api

import (
	"errors"
	"net/http"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
	"github.com/labstack/echo/v4"
)

// responseError process a correct response.
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

// responseError process an error response.
func responseError(err error) error {
	var vterror *vterrors.TrackingError

	if errors.As(err, &vterror) {
		return echo.NewHTTPError(vterror.HttpStatusCode, vterror.Error())
	}
	return err
}

// GetTicket returns a JSON containing a specific ticket.
// @Summary Return a ticket information.
// @Description return a ticket information.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func (api *API) GetTicket(c echo.Context) error {
	teamId := c.Param("team_id")
	id := c.Param("id")

	// Get the server and the configuration for the teamId.
	configuration, err := api.ticketServerStorage.ProjectConfig(teamId)
	if err != nil {
		return err
	}

	serverName := configuration.ServerName

	ticket, err := api.trackingServers[serverName].GetTicket(id)
	if err != nil {
		return responseError(err)
	}
	ticket.TeamID = teamId

	return response(c, http.StatusOK, ticket, "ticket")
}

// CreateTicket creates a ticket and returns a JSON containing the new ticket.
func (api *API) CreateTicket(c echo.Context) error {
	teamId := c.Param("team_id")
	ticket := new(model.Ticket)

	if err := c.Bind(ticket); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Get the server and the configuration for the teamId.
	configuration, err := api.ticketServerStorage.ProjectConfig(teamId)
	if err != nil {
		return err
	}

	// Retrieve the necessary values to create a ticket.
	ticket.TeamID = teamId
	ticket.Project = configuration.Project
	ticket.TicketType = configuration.VulnerabilityIssueType
	serverName := configuration.ServerName

	// Check if the ticket exists.
	var findingTicket model.FindingTicket
	findingTicket, err = api.storage.GetFindingTicket(ticket.FindingID, ticket.TeamID)
	if findingTicket.ID != "" {
		return echo.NewHTTPError(http.StatusConflict, "The ticket for this finding and team already exists.")
	}

	// Create the ticket in the tracker tool.
	ticket, err = api.trackingServers[serverName].CreateTicket(ticket)
	if err != nil {
		return responseError(err)
	}

	// Store the ticket created in the database.
	_, err = api.storage.CreateFindingTicket(*ticket)
	if err != nil {
		return responseError(err)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}

// FixTicket updates a ticket until a "done" state and returns a JSON containing the new ticket.
func (api *API) FixTicket(c echo.Context) error {
	teamId := c.Param("team_id")
	id := c.Param("id")

	// Get the server and the configuration for the teamId.
	configuration, err := api.ticketServerStorage.ProjectConfig(teamId)
	if err != nil {
		return err
	}

	serverName := configuration.ServerName

	ticket, err := api.trackingServers[serverName].FixTicket(id, configuration.FixedWorkflow)
	if err != nil {
		return responseError(err)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}

type WontFixForm struct {
	Reason string `json:"reason"`
}

// WontFixTicket updates a ticket until a "done" but with a won't fix reason state
// and returns a JSON containing the new ticket.
func (api *API) WontFixTicket(c echo.Context) error {
	teamId := c.Param("team_id")
	id := c.Param("id")
	form := new(WontFixForm)

	if err := c.Bind(form); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Get the server and the configuration for the teamId.
	configuration, err := api.ticketServerStorage.ProjectConfig(teamId)
	if err != nil {
		return err
	}

	serverName := configuration.ServerName

	ticket, err := api.trackingServers[serverName].WontFixTicket(id, configuration.WontFixWorkflow, form.Reason)
	if err != nil {
		return responseError(err)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}

// GetFindingTicket checks if a ticket was created and retrieves it if it is found.
func (api *API) GetFindingTicket(c echo.Context) error {
	teamId := c.Param("team_id")
	findingID := c.Param("finding_id")

	findingTicket, err := api.storage.GetFindingTicket(findingID, teamId)
	if err != nil {
		return err
	}
	return response(c, http.StatusOK, findingTicket, "ticket")
}
