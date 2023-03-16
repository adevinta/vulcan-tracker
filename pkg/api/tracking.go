/*
Copyright 2023 Adevinta
*/
package api

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// response process a correct response.
func response(c echo.Context, httpStatus int, data interface{}, dataType string) error {
	if data == nil {
		return c.NoContent(http.StatusNoContent)
	}

	resp := map[string]interface{}{}
	resp[dataType] = data

	return c.JSON(httpStatus, resp)
}

// responseError process an error response.
func responseError(err error) error {
	var vterror *vterrors.TrackingError

	if errors.As(err, &vterror) {
		return echo.NewHTTPError(vterror.HTTPStatusCode, vterror.Error())
	}
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return err
}

func isValidTeam(team string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(team)
}

// Healthcheck performs a simple query and returns an OK response.
func (api *API) Healthcheck(c echo.Context) error {
	err := api.storage.Healthcheck()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

// GetTicket returns a JSON containing a specific ticket.
func (api *API) GetTicket(c echo.Context) error {
	teamID := c.Param("team_id")
	id := c.Param("id")

	// Check if the team is an uuid
	if !isValidTeam(teamID) {
		return responseError(
			&vterrors.TrackingError{
				Msg:            "the team id should be a UUID",
				HTTPStatusCode: http.StatusBadRequest,
			})
	}

	// Get a ticket tracker client.
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServer, teamID, c.Logger())
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

	// Check if the team is an uuid
	if !isValidTeam(teamID) {
		return responseError(
			&vterrors.TrackingError{
				Msg:            "the team id should be a UUID",
				HTTPStatusCode: http.StatusBadRequest,
			})
	}

	if err := c.Bind(ticket); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Get the server and the configuration for the teamID.
	configuration, err := api.ticketServer.ProjectConfigByTeamID(teamID)
	if err != nil {
		return responseError(err)
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
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServer, teamID, c.Logger())
	if err != nil {
		return responseError(err)
	}

	// Create the ticket in the tracker tool.
	*ticket, err = ttClient.CreateTicket(*ticket)
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

	// Check if the team is an uuid
	if !isValidTeam(teamID) {
		return responseError(
			&vterrors.TrackingError{
				Msg:            "the team id should be a UUID",
				HTTPStatusCode: http.StatusBadRequest,
			})
	}

	// Get the server and the configuration for the teamID.
	configuration, err := api.ticketServer.ProjectConfigByTeamID(teamID)
	if err != nil {
		return responseError(err)
	}

	// Get a ticket tracker client.
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServer, teamID, c.Logger())
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

	// Check if the team is an uuid
	if !isValidTeam(teamID) {
		return responseError(
			&vterrors.TrackingError{
				Msg:            "the team id should be a UUID",
				HTTPStatusCode: http.StatusBadRequest,
			})
	}

	form := new(WontFixForm)

	if err := c.Bind(form); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	// Get the server and the configuration for the teamID.
	configuration, err := api.ticketServer.ProjectConfigByTeamID(teamID)
	if err != nil {
		return responseError(err)
	}

	// Get a ticket tracker client.
	ttClient, err := api.ticketTrackerBuilder.GenerateTicketTrackerClient(api.ticketServer, teamID, c.Logger())
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

	// Check if the team is an uuid
	if !isValidTeam(teamID) {
		return responseError(
			&vterrors.TrackingError{
				Msg:            "the team id should be a UUID",
				HTTPStatusCode: http.StatusBadRequest,
			})
	}

	findingTicket, err := api.storage.GetFindingTicket(findingID, teamID)
	if err != nil {
		return responseError(err)
	}
	return response(c, http.StatusOK, findingTicket, "ticket")
}
