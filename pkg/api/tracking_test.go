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

type mockTicketTrackerClient struct {
	tracking.TicketTracker
	project model.ProjectConfig
	tickets map[string]model.Ticket
}

type ticketExpected struct {
	Ticket model.Ticket `json:"ticket"`
}

func generateTicketExpected(ticketID string, tickets map[string]model.Ticket) []byte {
	ticket, err := json.Marshal(ticketExpected{Ticket: tickets[ticketID]})
	if err != nil {
		panic(err)
	}
	return ticket
}

func (mtt *mockTicketTrackerClient) GetTicket(id string) (model.Ticket, error) {
	value, ok := mtt.tickets[id]
	if ok {
		return value, nil
	}
	return model.Ticket{}, &vterrors.TrackingError{
		Err:            errors.New("ticket not found"),
		HTTPStatusCode: http.StatusNotFound,
	}
}

func (mtt *mockTicketTrackerClient) CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	ticket.Key = fmt.Sprintf("%s-%d", ticket.Project, len(mtt.tickets)+1)
	mtt.tickets[ticket.Key] = ticket
	return ticket, nil
}

type mockTicketServer struct {
	tracking.TicketServer
	projects map[string]model.ProjectConfig
}

func (ms *mockTicketServer) ProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {
	value, ok := ms.projects[teamID]
	if ok {
		return value, nil
	}
	return model.ProjectConfig{}, &vterrors.TrackingError{
		Err:            errors.New("project not found"),
		HTTPStatusCode: http.StatusNotFound,
	}
}

func (ms *mockTicketServer) ServerConf(serverID string) (model.TrackerConfig, error) {
	return model.TrackerConfig{
		ID: serverID,
	}, nil
}

type mockTicketTrackerBuilder struct {
	tickets  map[string]model.Ticket
	projects map[string]model.ProjectConfig
}

func (mttb *mockTicketTrackerBuilder) GenerateTicketTrackerClient(_ tracking.TicketServer,
	teamID string, _ echo.Logger) (tracking.TicketTracker, error) {
	value, ok := mttb.projects[teamID]
	if !ok {
		return &mockTicketTrackerClient{}, &vterrors.TrackingError{
			Err:            errors.New("project not found"),
			HTTPStatusCode: http.StatusNotFound,
		}
	}

	return &mockTicketTrackerClient{
		project: value,
		tickets: mttb.tickets,
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

func TestGetTicket(t *testing.T) {
	ticket := model.Ticket{
		ID:          "1000",
		Key:         "TEST-1",
		TeamID:      "80287cf6-db31-47f8-a9aa-792f214b1f88",
		Summary:     "Summary TEST-1",
		Description: "Description TEST-1",
		Project:     "TEST",
		Status:      ToDo,
		TicketType:  "Vulnerability",
		FindingID:   "FindingID1",
	}

	t1 := ticket // make a copy of the ticket.
	tickets := map[string]model.Ticket{
		t1.Key: t1,
	}

	project := model.ProjectConfig{
		ID:                     "projectID1",
		Name:                   "ProjectName1",
		TeamID:                 "80287cf6-db31-47f8-a9aa-792f214b1f88",
		TrackerConfigID:        "JiraServerID",
		Project:                "TEST",
		VulnerabilityIssueType: "Vulnerability",
	}

	p1 := project // make a copy of the project.
	projects := map[string]model.ProjectConfig{
		p1.TeamID: p1,
	}

	tests := []struct {
		name           string
		teamID         string
		ticketID       string
		api            *API
		wantStatusCode int
		wantError      *echo.HTTPError
		wantTicket     string
	}{
		{
			name:     "HappyPath",
			teamID:   "80287cf6-db31-47f8-a9aa-792f214b1f88",
			ticketID: "TEST-1",
			api: &API{
				ticketServer: &mockTicketServer{
					projects: projects,
				},
				ticketTrackerBuilder: &mockTicketTrackerBuilder{
					tickets:  tickets,
					projects: projects,
				},
				storage: &mockStorage{},
			},
			wantStatusCode: http.StatusOK,
			wantTicket:     "TEST-1",
		},
		{
			name:     "TicketNotFound",
			teamID:   "80287cf6-db31-47f8-a9aa-792f214b1f88",
			ticketID: "TEST-NOTFOUND",
			api: &API{
				ticketServer: &mockTicketServer{
					projects: projects,
				},
				ticketTrackerBuilder: &mockTicketTrackerBuilder{
					tickets:  tickets,
					projects: projects,
				},
				storage: &mockStorage{},
			},
			wantError: &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: "ticket not found",
			},
		},
		{
			name:     "TeamNotFound",
			teamID:   "c99fed50-c612-4f99-863f-6d3274e2b2b6",
			ticketID: "TEST-1",
			api: &API{
				ticketServer: &mockTicketServer{
					projects: projects,
				},
				ticketTrackerBuilder: &mockTicketTrackerBuilder{
					tickets:  tickets,
					projects: projects,
				},
				storage: &mockStorage{},
			},
			wantError: &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: "project not found",
			},
		},
		{
			name:     "TeamNotUUID",
			teamID:   "teamNotUUID",
			ticketID: "TEST-TICKET-NOT-UUID",
			api: &API{
				ticketServer: &mockTicketServer{
					projects: projects,
				},
				ticketTrackerBuilder: &mockTicketTrackerBuilder{
					tickets:  tickets,
					projects: projects,
				},
				storage: &mockStorage{},
			},
			wantError: &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "the team id should be a UUID",
			},
		},
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

			err := tt.api.GetTicket(c)
			if err != nil { // When the endpoint returns an error.
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

				clientBuilder := tt.api.ticketTrackerBuilder.(*mockTicketTrackerBuilder)
				diff = cmp.Diff(rec.Body.String(), string(generateTicketExpected(tt.wantTicket, clientBuilder.tickets))+"\n")
				if diff != "" {
					t.Fatalf("the body content does not match the expected one. diff: %v\n", diff)
				}
			}
		})
	}
}

func TestCreateTicket(t *testing.T) {
	ticket := model.Ticket{
		ID:          "1000",
		Key:         "TEST-1",
		TeamID:      "80287cf6-db31-47f8-a9aa-792f214b1f88",
		Summary:     "Summary TEST-1",
		Description: "Description TEST-1",
		Project:     "TEST",
		Status:      ToDo,
		TicketType:  "Vulnerability",
		FindingID:   "FindingID1",
	}

	t1 := ticket // make a copy of the ticket.
	tickets := map[string]model.Ticket{
		t1.Key: t1,
	}

	project := model.ProjectConfig{
		ID:                     "projectID1",
		Name:                   "ProjectName1",
		TeamID:                 "80287cf6-db31-47f8-a9aa-792f214b1f88",
		TrackerConfigID:        "JiraServerID",
		Project:                "TEST",
		VulnerabilityIssueType: "Vulnerability",
	}

	p1 := project // make a copy of the project.
	projects := map[string]model.ProjectConfig{
		p1.TeamID: p1,
	}

	tests := []struct {
		name           string
		teamID         string
		ticket         model.Ticket
		api            *API
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
			api: &API{
				ticketServer: &mockTicketServer{
					projects: projects,
				},
				ticketTrackerBuilder: &mockTicketTrackerBuilder{
					tickets:  tickets,
					projects: projects,
				},
				storage: &mockStorage{},
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
			api: &API{
				ticketServer: &mockTicketServer{
					projects: projects,
				},
				ticketTrackerBuilder: &mockTicketTrackerBuilder{
					tickets:  tickets,
					projects: projects,
				},
				storage: &mockStorage{},
			},
			wantError: &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: "project not found",
			},
		},
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

			err = tt.api.CreateTicket(c)

			if err != nil { // When the endpoint returns an error.
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

				clientBuilder := tt.api.ticketTrackerBuilder.(*mockTicketTrackerBuilder)
				diff = cmp.Diff(
					rec.Body.String(), string(generateTicketExpected(tt.wantTicket, clientBuilder.tickets))+"\n")
				if diff != "" {
					t.Fatalf("the body content does not match the expected one. diff: %v\n", diff)
				}
			}
		})
	}
}
