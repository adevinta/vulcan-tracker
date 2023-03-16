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
	project model.ProjectConfig
}

var (
	h        *API
	tickets  map[string]*model.Ticket
	projects map[string]*model.ProjectConfig
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

func (mtt *mockTicketTracker) GetTicket(id string) (model.Ticket, error) {
	value, ok := tickets[id]
	if ok {
		return *value, nil
	}
	return model.Ticket{}, &vterrors.TrackingError{
		Msg:            "ticket not found",
		HTTPStatusCode: http.StatusNotFound,
	}
}

func (mtt *mockTicketTracker) CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	ticket.Key = fmt.Sprintf("%s-%d", ticket.Project, len(tickets)+1)
	tickets[ticket.Key] = &ticket
	return ticket, nil
}

type mockTicketServer struct {
	tracking.TicketServer
}

func (ms *mockTicketServer) ProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {
	value, ok := projects[teamID]
	if ok {
		return *value, nil
	}
	return model.ProjectConfig{}, &vterrors.TrackingError{
		Msg:            "project not found",
		HTTPStatusCode: http.StatusNotFound,
	}

}

func (ms *mockTicketServer) ServerConf(serverID string) (model.TrackerConfig, error) {
	return model.TrackerConfig{
		ID: serverID,
	}, nil
}

type mockTicketTrackerBuilder struct {
}

func (mttb *mockTicketTrackerBuilder) GenerateTicketTrackerClient(_ tracking.TicketServer,
	teamID string, _ echo.Logger) (tracking.TicketTracker, error) {
	value, ok := projects[teamID]
	if !ok {
		return &mockTicketTracker{}, &vterrors.TrackingError{
			Msg:            "project not found",
			HTTPStatusCode: http.StatusNotFound,
		}
	}

	return &mockTicketTracker{
		project: *value,
	}, nil
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
	return model.FindingTicket{}, errors.New("finding ticket not found")
}

func setupSubTest(t *testing.T) {
	t.Log("setup sub test")

	tickets = map[string]*model.Ticket{
		"TEST-1": {
			ID:          "1000",
			Key:         "TEST-1",
			TeamID:      "80287cf6-db31-47f8-a9aa-792f214b1f88",
			Summary:     "Summary TEST-1",
			Description: "Description TEST-1",
			Project:     "TEST",
			Status:      ToDo,
			TicketType:  "Vulnerability",
			FindingID:   "FindingID1",
		},
	}

	projects = make(map[string]*model.ProjectConfig)
	projects["80287cf6-db31-47f8-a9aa-792f214b1f88"] = &model.ProjectConfig{
		ID:                     "projectID1",
		Name:                   "ProjectName1",
		TeamID:                 "80287cf6-db31-47f8-a9aa-792f214b1f88",
		ServerID:               "JiraServerID",
		Project:                "TEST",
		VulnerabilityIssueType: "Vulnerability",
		FixedWorkflow:          nil,
		WontFixWorkflow:        nil,
		AutoCreate:             false,
	}

	h = &API{
		ticketServer:         &mockTicketServer{},
		ticketTrackerBuilder: &mockTicketTrackerBuilder{},
		storage:              &mockStorage{},
	}
}

func TestGetTicket(t *testing.T) {

	tests := []struct {
		name           string
		teamID         string
		ticketID       string
		wantStatusCode int
		wantError      *echo.HTTPError
		wantTicket     string
	}{
		{
			name:           "HappyPath",
			teamID:         "80287cf6-db31-47f8-a9aa-792f214b1f88",
			ticketID:       "TEST-1",
			wantStatusCode: http.StatusOK,
			wantTicket:     "TEST-1",
		},
		{
			name:     "TicketNotFound",
			teamID:   "80287cf6-db31-47f8-a9aa-792f214b1f88",
			ticketID: "TEST-NOTFOUND",
			wantError: &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: "ticket not found",
			},
		},
		{
			name:     "TeamNotFound",
			teamID:   "c99fed50-c612-4f99-863f-6d3274e2b2b6",
			ticketID: "TEST-1",
			wantError: &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: "project not found",
			},
		},
		{
			name:     "TeamNotUUID",
			teamID:   "teamNotUUID",
			ticketID: "TEST-TICKET-NOT-UUID",
			wantError: &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "the team id should be a UUID",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t)

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
					diff := cmp.Diff(he, tt.wantError)
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

func TestCreateTicket(t *testing.T) {
	tests := []struct {
		name           string
		teamID         string
		ticket         model.Ticket
		wantStatusCode int
		wantTicket     string
		wantError      *echo.HTTPError
	}{
		{
			name:   "HappyPath",
			teamID: "80287cf6-db31-47f8-a9aa-792f214b1f88",
			ticket: model.Ticket{
				Summary:     "Summary TEST-2",
				Description: "Description TEST-2",
				FindingID:   "12345678",
			},
			wantStatusCode: http.StatusOK,
			wantTicket:     "TEST-2",
		},
		{
			name:   "TeamNotFound",
			teamID: "c99fed50-c612-4f99-863f-6d3274e2b2b6",
			ticket: model.Ticket{
				Summary:     "Summary TEST-2",
				Description: "Description TEST-2",
				FindingID:   "12345678",
			},
			wantError: &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: "project not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t)

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
					diff := cmp.Diff(he, tt.wantError)
					if diff != "" {
						t.Fatalf("the status code does not match the expected one. diff: %v\n", diff)
					}
				}
			} else {
				diff := cmp.Diff(rec.Code, tt.wantStatusCode)
				if diff != "" {
					t.Fatalf("the status code does not match the expected one. diff: %v\n", diff)
				}

				diff = cmp.Diff(
					rec.Body.String(), string(generateTicketExpected(tt.wantTicket))+"\n")
				if diff != "" {
					t.Fatalf("the body content does not match the expected one. diff: %v\n", diff)
				}
			}
		})
	}
}
