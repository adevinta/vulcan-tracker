/*
Copyright 2022 Adevinta
*/
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"

	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"
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

func generateTicketExpected(ticketID string) []byte {
	type ticketExpected struct {
		Ticket model.Ticket `json:"ticket"`
	}

	ticket, err := json.Marshal(ticketExpected{Ticket: *tickets[ticketID]})
	if err != nil {
		panic(err)
	}
	return ticket
}

func (mtt *mockTicketTracker) GetTicket(id string) (*model.Ticket, error) {
	value, ok := tickets[id]
	if ok {
		return value, nil
	}
	return nil, &vterrors.TrackingError{
		Msg:            "ticket not found",
		HttpStatusCode: http.StatusNotFound,
	}
}

type mockStorage struct {
	storage.TicketServerStorage
}

func (ms *mockStorage) ProjectConfigByTeamID(teamId string) (*model.ProjectConfig, error) {
	return &model.ProjectConfig{
		ServerID: "test_server",
	}, nil
}

func (ms *mockStorage) ServerConf(name string) (*model.TrackerConfig, error) {
	return &model.TrackerConfig{
		ID: "ServerID",
	}, nil
}

type mockTicketTrackerBuilder struct {
}

func (mttb *mockTicketTrackerBuilder) GenerateTicketTrackerClient(storage storage.TicketServerStorage,
	teamId string, logger echo.Logger) (tracking.TicketTracker, error) {
	return &mockTicketTracker{}, nil

}

func TestGetTicket(t *testing.T) {

	tests := []struct {
		name           string
		teamID         string
		ticketID       string
		wantStatusCode int
		wantTicket     []byte
	}{
		{
			name:           "HappyPath",
			teamID:         "test_server",
			ticketID:       "TEST-1",
			wantStatusCode: http.StatusOK,
			wantTicket:     generateTicketExpected("TEST-1"),
		},
		{
			name:           "NotFound",
			teamID:         "test_server",
			ticketID:       "TEST-NOTFOUND",
			wantStatusCode: http.StatusNotFound,
		},
	}

	h := &API{
		ticketServerStorage:  &mockStorage{},
		ticketTrackerBuilder: &mockTicketTrackerBuilder{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:team_id/tickets/:id")
			c.SetParamNames("team_id", "id")
			c.SetParamValues(tt.teamID, tt.ticketID)

			err := h.GetTicket(c)
			if err != nil { // When the endpoint returns a error
				he, ok := err.(*echo.HTTPError)
				if ok {
					diff := cmp.Diff(he.Code, tt.wantStatusCode)
					if diff != "" {
						t.Fatalf("the status code does not match the expected one. diff: %v\n", diff)
					}
				}
			} else {
				diff := cmp.Diff(rec.Code, tt.wantStatusCode)
				if diff != "" {
					t.Fatalf("the status code does not match the expected one. diff: %v\n", diff)
				}

				diff = cmp.Diff(rec.Body.String(), string(tt.wantTicket)+"\n")
				if diff != "" {
					t.Fatalf("the body content does not match the expected one. diff: %v\n", diff)
				}
			}
		})
	}
}
