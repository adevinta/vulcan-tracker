/*
Copyright 2022 Adevinta
*/
package jira

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
)

const (
	ToDo       string = "To Do"
	InProgress string = "In Progress"
	Resolved   string = "Resolved"
)

type mockJiraClient struct {
	TicketTrackingClient
	tickets map[string]model.Ticket
}

type mockLogger struct {
	echo.Logger
}

func (l *mockLogger) Errorf(_ string, _ ...interface{}) {
	// Do nothing logger
}

func (mj *mockJiraClient) GetTicket(id string) (model.Ticket, error) {
	value, ok := mj.tickets[id]
	if ok {
		return value, nil
	}
	return model.Ticket{}, &vterrors.TrackingError{
		Err:            fmt.Errorf("ticket %s not found", id),
		HTTPStatusCode: http.StatusNotFound,
	}
}

func (mj *mockJiraClient) FindTicket(projectKey, vulnerabilityIssueType, text string) (model.Ticket, error) {

	var ticketsFound []model.Ticket

	for _, ticket := range mj.tickets {
		if ticket.Project != projectKey {
			continue
		}
		if ticket.TicketType != vulnerabilityIssueType {
			continue
		}
		if strings.Contains(ticket.Description, text) {
			ticketsFound = append(ticketsFound, ticket)
		}
	}
	if len(ticketsFound) == 0 {
		return model.Ticket{}, nil
	}
	return ticketsFound[0], nil
}

func (mj *mockJiraClient) CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	ticket.Key = fmt.Sprintf("%s-%d", ticket.Project, len(mj.tickets)+1)
	ticket.ID = fmt.Sprintf("%d", 1000+len(mj.tickets)+1)
	ticket.Status = ToDo

	mj.tickets[ticket.Key] = ticket

	return ticket, nil
}

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}

func TestGetTicket(t *testing.T) {
	t1 := model.Ticket{
		ID:          "1000",
		Key:         "TEST-1",
		TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
		FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
		Summary:     "Summary TEST-1",
		Description: GenerateDescriptionText("Description TEST-1", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
		Project:     "TEST",
		Status:      ToDo,
		TicketType:  "Vulnerability",
		Labels:      []string{"Vulnerability"},
	}
	t2 := model.Ticket{
		ID:          "1001",
		Key:         "TEST-2",
		TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
		FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
		Summary:     "Summary TEST-2",
		Description: GenerateDescriptionText("Description TEST-2", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
		Project:     "TEST",
		Status:      InProgress,
		TicketType:  "Vulnerability",
		Labels:      []string{"Potential Vulnerability"},
	}

	tickets := map[string]model.Ticket{
		t1.Key: t1,
		t2.Key: t2,
	}

	tests := []struct {
		name             string
		ticketID         string
		jiraTicketClient *TrackerClient
		want             model.Ticket
		wantErr          error
	}{
		{
			name:     "HappyPath",
			ticketID: "TEST-1",
			jiraTicketClient: &TrackerClient{
				Client: &mockJiraClient{
					tickets: tickets,
				},
				Logger: &mockLogger{},
			},
			want: model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-1",
				Description: GenerateDescriptionText("Description TEST-1", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      ToDo,
				TicketType:  "Vulnerability",
				Labels:      []string{"Vulnerability"},
				URLTracker:  "/browse/TEST-1",
			},
			wantErr: nil,
		},
		{
			name:     "KeyNotFound",
			ticketID: "NOTFOUND",
			jiraTicketClient: &TrackerClient{
				Client: &mockJiraClient{
					tickets: tickets,
				},
				Logger: &mockLogger{},
			},
			want:    model.Ticket{},
			wantErr: errors.New("ticket NOTFOUND not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.jiraTicketClient.GetTicket(tt.ticketID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Fatalf("ticket does not match expected one. diff: %s\n", diff)
			}
		})
	}
}

func TestCreateTicket(t *testing.T) {
	t1 := model.Ticket{
		ID:          "1000",
		Key:         "TEST-1",
		TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
		FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
		Summary:     "Summary TEST-1",
		Description: GenerateDescriptionText("Description TEST-1", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
		Project:     "TEST",
		Status:      ToDo,
		TicketType:  "Vulnerability",
		Labels:      []string{"Vulnerability"},
	}
	t2 := model.Ticket{
		ID:          "1001",
		Key:         "TEST-2",
		TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
		FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
		Summary:     "Summary TEST-2",
		Description: GenerateDescriptionText("Description TEST-2", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
		Project:     "TEST",
		Status:      InProgress,
		TicketType:  "Vulnerability",
		Labels:      []string{"Potential Vulnerability"},
	}

	tickets := map[string]model.Ticket{
		t1.Key: t1,
		t2.Key: t2,
	}

	tests := []struct {
		name             string
		newTicket        model.Ticket
		jiraTicketClient *TrackerClient
		want             model.Ticket
		wantErr          error
	}{
		{
			name: "HappyPath",
			newTicket: model.Ticket{
				TeamID:      "11c2c999-14a4-434f-ab02-107e2cc14324",
				FindingID:   "fcafb55e-f8d4-4a10-b073-149b84554b94",
				Summary:     "Summary New Ticket",
				Description: "Description New Ticket",
				Project:     "TEST",
				TicketType:  "Vulnerability",
			},
			jiraTicketClient: &TrackerClient{
				Client: &mockJiraClient{
					tickets: tickets,
				},
				Logger: &mockLogger{},
			},
			want: model.Ticket{
				TeamID:      "11c2c999-14a4-434f-ab02-107e2cc14324",
				FindingID:   "fcafb55e-f8d4-4a10-b073-149b84554b94",
				Summary:     "Summary New Ticket",
				Description: GenerateDescriptionText("Description New Ticket", "fcafb55e-f8d4-4a10-b073-149b84554b94", "11c2c999-14a4-434f-ab02-107e2cc14324"),
				Project:     "TEST",
				Status:      ToDo,
				TicketType:  "Vulnerability",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.jiraTicketClient.CreateTicket(tt.newTicket)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(model.Ticket{}, "ID", "Key", "URLTracker"))
			if diff != "" {
				t.Fatalf("ticket does not match expected one. diff: %s\n", diff)
			}
		})
	}
}
