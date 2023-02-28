/*
Copyright 2023 Adevinta
*/
package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
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
		return echo.NewHTTPError(vterror.HTTPStatusCode, vterror.Error())
	}
	return err
}

// GetTicket returns a JSON containing a specific ticket.
func (api *API) GetTicket(c echo.Context) error {
	teamID := c.Param("team_id")
	id := c.Param("id")

	// Get a ticket tracker client.
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServerStorage, teamID, c.Logger())
	if err != nil {
		return responseError(err)
	}
	ticket, err := ttClient.GetTicket(id)
	if err != nil {
		return responseError(err)
	}
	ticket.TeamID = teamID

	return response(c, http.StatusOK, ticket, "ticket")
}

// CreateTicket creates a ticket and returns a JSON containing the new ticket.
func (api *API) CreateTicket(c echo.Context) error {
	teamID := c.Param("team_id")
	ticket := new(model.Ticket)

	if err := c.Bind(ticket); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Get the server and the configuration for the teamID.
	configuration, err := api.ticketServerStorage.ProjectConfigByTeamID(teamID)
	if err != nil {
		return err
	}

	// Retrieve the necessary values to create a ticket.
	ticket.TeamID = teamID
	ticket.Project = configuration.Project
	ticket.TicketType = configuration.VulnerabilityIssueType

	// Check if the ticket exists.
	findingTicket, err := api.storage.GetFindingTicket(ticket.FindingID, ticket.TeamID)
	if findingTicket.ID != "" {
		return echo.NewHTTPError(http.StatusConflict, "The ticket for this finding and team already exists.")
	}

	// Get a ticket tracker client.
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServerStorage, teamID, c.Logger())
	if err != nil {
		return responseError(err)
	}

	// Create the ticket in the tracker tool.
	ticket, err = ttClient.CreateTicket(ticket)
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
	teamID := c.Param("team_id")
	id := c.Param("id")

	// Get the server and the configuration for the teamID.
	configuration, err := api.ticketServerStorage.ProjectConfigByTeamID(teamID)
	if err != nil {
		return err
	}

	// Get a ticket tracker client.
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServerStorage, teamID, c.Logger())
	if err != nil {
		return responseError(err)
	}

	ticket, err := ttClient.FixTicket(id, configuration.FixedWorkflow)
	if err != nil {
		return responseError(err)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}

// WontFixForm represents the information associated to a ticket that will be marked as Won't Fix.
type WontFixForm struct {
	Reason string `json:"reason"`
}

// WontFixTicket updates a ticket until a "done" but with a won't fix reason state
// and returns a JSON containing the new ticket.
func (api *API) WontFixTicket(c echo.Context) error {
	teamID := c.Param("team_id")
	id := c.Param("id")
	form := new(WontFixForm)

	if err := c.Bind(form); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Get the server and the configuration for the teamID.
	configuration, err := api.ticketServerStorage.ProjectConfigByTeamID(teamID)
	if err != nil {
		return err
	}

	// Get a ticket tracker client.
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServerStorage, teamID, c.Logger())
	if err != nil {
		return responseError(err)
	}
	ticket, err := ttClient.WontFixTicket(id, configuration.WontFixWorkflow, form.Reason)
	if err != nil {
		return responseError(err)
	}

	return response(c, http.StatusOK, ticket, "ticket")
}

// GetFindingTicket checks if a ticket was created and retrieves it if it is found.
func (api *API) GetFindingTicket(c echo.Context) error {
	teamID := c.Param("team_id")
	findingID := c.Param("finding_id")

	findingTicket, err := api.storage.GetFindingTicket(findingID, teamID)
	if err != nil {
		return err
	}
	return response(c, http.StatusOK, findingTicket, "ticket")
}
