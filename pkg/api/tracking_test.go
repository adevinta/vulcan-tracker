/*
Copyright 2022 Adevinta
*/
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	tickets = map[string]model.Ticket{
		"TEST-1": {
			ID:          "1000",
			Key:         "TEST-1",
			TeamID:      "projectTeamID",
			Summary:     "Summary TEST-1",
			Description: "Description TEST-1",
			Project:     "TEST",
			Status:      ToDo,
			TicketType:  "Vulnerability",
			FindingID:   "FindingID1",
		},
	}
)

func generateTicketExpected(ticketID string) []byte {
	type ticketExpected struct {
		Ticket model.Ticket `json:"ticket"`
	}

	ticket, err := json.Marshal(ticketExpected{Ticket: tickets[ticketID]})
	if err != nil {
		panic(err)
	}
	return ticket
}

func (mtt *mockTicketTracker) GetTicket(id string) (model.Ticket, error) {
	value, ok := tickets[id]
	if ok {
		return value, nil
	}
	return model.Ticket{}, &vterrors.TrackingError{
		Msg:            "ticket not found",
		HTTPStatusCode: http.StatusNotFound,
	}
}

func (mtt *mockTicketTracker) CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	ticket.Key = fmt.Sprintf("%s-%d", ticket.Project, len(tickets)+1)
	tickets[ticket.Key] = ticket
	return ticket, nil
}

type mockTicketServer struct {
	tracking.TicketServer
}

func (ms *mockTicketServer) ProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {
	return model.ProjectConfig{
		ServerID:               "test_server",
		TeamID:                 teamID,
		Project:                "TEST",
		VulnerabilityIssueType: "Vulnerability",
	}, nil
}

func (ms *mockTicketServer) ServerConf(serverID string) (model.TrackerConfig, error) {
	return model.TrackerConfig{
		ID: serverID,
	}, nil
}

type mockTicketTrackerBuilder struct {
}

func (mttb *mockTicketTrackerBuilder) GenerateTicketTrackerClient(_ tracking.TicketServer,
	_ string, _ echo.Logger) (tracking.TicketTracker, error) {
	return &mockTicketTracker{}, nil
}

type mockStorage struct {
	storage.Storage
	findingTickets []model.FindingTicket
}

func (ms *mockStorage) CreateFindingTicket(ticket model.Ticket) (model.FindingTicket, error) {
	ft := model.FindingTicket{
		ID:         ticket.ID,
		FindingID:  ticket.FindingID,
		TeamID:     ticket.TeamID,
		URLTracker: ticket.URLTracker,
	}
	ms.findingTickets = append(ms.findingTickets, ft)

	return ft, nil
}

func (ms *mockStorage) GetFindingTicket(findingID, teamID string) (model.FindingTicket, error) {
	for _, ft := range ms.findingTickets {
		if ft.FindingID == findingID && ft.TeamID == teamID {
			return ft, nil
		}
	}
	return model.FindingTicket{}, errors.New("Finding ticket not found")
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
			teamID:         "projectTeamID",
			ticketID:       "TEST-1",
			wantStatusCode: http.StatusOK,
			wantTicket:     generateTicketExpected("TEST-1"),
		},
		{
			name:           "NotFound",
			teamID:         "projectTeamID",
			ticketID:       "TEST-NOTFOUND",
			wantStatusCode: http.StatusNotFound,
		},
	}

	h := &API{
		ticketServer:         &mockTicketServer{},
		ticketTrackerBuilder: &mockTicketTrackerBuilder{},
		storage:              &mockStorage{},
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

func TestCreateTicket(t *testing.T) {
	tests := []struct {
		name           string
		teamID         string
		ticket         model.Ticket
		wantStatusCode int
		wantTicket     string
	}{
		{
			name:   "HappyPath",
			teamID: "projectTeamID",
			ticket: model.Ticket{
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
				FindingID:   "12345678",
			},
			wantStatusCode: http.StatusOK,
			wantTicket:     "TEST-2",
		},
	}

	h := &API{
		ticketServer:         &mockTicketServer{},
		ticketTrackerBuilder: &mockTicketTrackerBuilder{},
		storage:              &mockStorage{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			qsStrData, err := json.Marshal(tt.ticket)
			if err != nil {
				t.Fatalf("Should be able to marshal the fixture : %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(qsStrData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			c.SetPath("/:team_id/tickets")
			c.SetParamNames("team_id")
			c.SetParamValues(tt.teamID)

			err = h.CreateTicket(c)

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

				diff = cmp.Diff(rec.Body.String(), string(generateTicketExpected(tt.wantTicket))+"\n")
				if diff != "" {
					t.Fatalf("the body content does not match the expected one. diff: %v\n", diff)
				}
			}
		})
	}
}
