/*
Copyright 2022 Adevinta
*/
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
)

const (
	ToDo       string = "To Do"
	InProgress string = "In Progress"
	Resolved   string = "Resolved"
)

type mockTicketTracker struct {
	tracking.TicketTracker
}

var (
	tickets = map[string]*model.Ticket{
		"TEST-1": {
			ID:          "1000",
			Key:         "TEST-1",
			TeamID:      "test_server",
			Summary:     "Summary TEST-1",
			Description: "Description TEST-1",
			Project:     "TEST",
			Status:      ToDo,
			TicketType:  "Vulnerability",
		},
	}
)

func (mtt *mockTicketTracker) GetTicket(id string) (*model.Ticket, error) {
	value, ok := tickets[id]
	if ok {
		return value, nil
	}
	return nil, &vterrors.TrackingError{
		Err:            errors.New("ticket not found"),
		HttpStatusCode: http.StatusNotFound,
	}
}

type mockStorage struct {
	storage.TicketServerStorage
}

func (ms *mockStorage) ProjectConfig(teamId string) (*model.ProjectConfig, error) {
	return &model.ProjectConfig{
		ServerName: "test_server",
	}, nil
}

func TestGetTicket(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:team_id/tickets/:id")
	c.SetParamNames("team_id", "id")
	c.SetParamValues("test_server", "TEST-1")
	h := &API{
		trackingServers: map[string]tracking.TicketTracker{
			"test_server": &mockTicketTracker{},
		},
		ticketServerStorage: &mockStorage{},
	}

	type ticketExpected struct {
		Ticket model.Ticket `json:"ticket"`
	}

	out, err := json.Marshal(ticketExpected{Ticket: *tickets["TEST-1"]})
	if err != nil {
		panic(err)
	}
	if assert.NoError(t, h.GetTicket(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, string(out)+"\n", rec.Body.String())
	}

}

func TestGetTicket_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/:team_id/tickets/:id")
	c.SetParamNames("team_id", "id")
	c.SetParamValues("test_server", "TEST-NOTFOUND")
	h := &API{
		trackingServers: map[string]tracking.TicketTracker{
			"test_server": &mockTicketTracker{},
		},
		ticketServerStorage: &mockStorage{},
	}

	err := h.GetTicket(c)
	if assert.NotNil(t, err) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			assert.Equal(t, http.StatusNotFound, he.Code)
		}
	}

}
