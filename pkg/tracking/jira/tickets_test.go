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
	testutil "github.com/adevinta/vulcan-tracker/pkg/testutils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
)

const (
	ToDo       string = "To Do"
	InProgress string = "In Progress"
	Resolved   string = "Resolved"
)

type MockJiraClient struct {
	TicketTrackingClient
	tickets     map[string]*model.Ticket
	transitions map[string][]model.Transition
}

type mockLogger struct {
	echo.Logger
}

func (l *mockLogger) Errorf(_ string, _ ...interface{}) {
	// Do nothing logger
}

func (mj *MockJiraClient) GetTicket(id string) (model.Ticket, error) {
	value, ok := mj.tickets[id]
	if ok {
		return *value, nil
	}
	return model.Ticket{}, &vterrors.TrackingError{
		Msg:            fmt.Sprintf("ticket %s not found", id),
		HTTPStatusCode: http.StatusNotFound,
	}
}

func (mj *MockJiraClient) FindTicket(projectKey, vulnerabilityIssueType, text string) (model.Ticket, error) {

	var ticketsFound []model.Ticket

	for _, ticket := range mj.tickets {
		if ticket.Project != projectKey {
			continue
		}
		if ticket.TicketType != vulnerabilityIssueType {
			continue
		}
		if strings.Contains(ticket.Description, text) {
			ticketsFound = append(ticketsFound, *ticket)
		}
	}
	if len(ticketsFound) == 0 {
		return model.Ticket{}, nil
	}
	return ticketsFound[0], nil
}

func (mj *MockJiraClient) CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	ticket.Key = fmt.Sprintf("%s-%d", ticket.Project, len(mj.tickets)+1)
	ticket.ID = fmt.Sprintf("%d", 1000+len(mj.tickets)+1)
	ticket.Status = ToDo

	mj.tickets[ticket.Key] = &ticket

	return ticket, nil
}

func (mj *MockJiraClient) GetTicketTransitions(id string) ([]model.Transition, error) {
	tempTicket, err := mj.GetTicket(id)
	if err != nil {
		return nil, err
	}
	ticket, _ := mj.tickets[tempTicket.Key]
	transitions, ok := mj.transitions[ticket.Status]
	if ok {
		return transitions, nil
	}
	return nil, fmt.Errorf("transitions for %s not found", id)

}
func (mj *MockJiraClient) DoTransition(id, idTransition string) error {
	tempTicket, err := mj.GetTicket(id)
	if err != nil {
		return err
	}
	ticket, _ := mj.tickets[tempTicket.Key]
	transitions, ok := mj.transitions[ticket.Status]
	if !ok {
		return fmt.Errorf("transitions for %s not found", id)
	}

	for _, transition := range transitions {
		if transition.ID == idTransition {
			ticket.Status = transition.ToName
			transitionsDone = append(transitionsDone, ticket.Status)
			break
		}
	}
	return nil
}
func (mj *MockJiraClient) DoTransitionWithResolution(id, idTransition, resolution string) error {
	tempTicket, err := mj.GetTicket(id)
	if err != nil {
		return err
	}
	ticket, _ := mj.tickets[tempTicket.Key]
	transitions, ok := mj.transitions[ticket.Status]
	if !ok {
		return fmt.Errorf("transitions for %s not found", id)
	}

	for _, transition := range transitions {
		if transition.ID == idTransition {
			ticket.Status = transition.ToName
			ticket.Resolution = resolution
			transitionsDone = append(transitionsDone, ticket.Status)
			break
		}
	}
	return nil
}

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}

var (
	tc              *TC
	transitionsDone []string
	transitions     = map[string][]model.Transition{
		ToDo: {
			{
				ID:     "1001",
				ToName: InProgress,
			},
		},
		InProgress: {
			{
				ID:     "1002",
				ToName: Resolved,
			},
			{
				ID:     "1000",
				ToName: ToDo,
			},
		},
	}
	trResolveFromAnyStatus = map[string][]model.Transition{
		ToDo: {
			{
				ID:     "1001",
				ToName: InProgress,
			},
			{
				ID:     "1002",
				ToName: Resolved,
			},
		},
		InProgress: {
			{
				ID:     "1002",
				ToName: Resolved,
			},
			{
				ID:     "1000",
				ToName: ToDo,
			},
		},
	}
)

func setupSubTest(t *testing.T, transitions map[string][]model.Transition) {
	t.Log("setup sub test")

	tickets := make(map[string]*model.Ticket)
	tickets["TEST-1"] = &model.Ticket{
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
	tickets["TEST-2"] = &model.Ticket{
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

	jiraClient := MockJiraClient{
		tickets:     tickets,
		transitions: transitions,
	}
	tc = &TC{
		Client: &jiraClient,
		Logger: &mockLogger{},
	}
	transitionsDone = []string{}

}

func TestGetTicket(t *testing.T) {
	tests := []struct {
		name     string
		ticketID string
		want     model.Ticket
		wantErr  error
	}{
		{
			name:     "HappyPath",
			ticketID: "TEST-1",
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
			want:     model.Ticket{},
			wantErr:  errors.New("ticket NOTFOUND not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, transitions)

			got, err := tc.GetTicket(tt.ticketID)
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

func TestGetTicketTransition(t *testing.T) {
	tests := []struct {
		name     string
		ticketID string
		want     []model.Transition
		wantErr  error
	}{
		{
			name:     "HappyPath",
			ticketID: "TEST-1",
			want: []model.Transition{
				{
					ID:     "1001",
					ToName: InProgress,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, transitions)

			got, err := tc.GetTransitions(tt.ticketID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Fatalf("transition does not match expected one. diff: %s\n", diff)
			}
		})
	}
}

func TestCreateTicket(t *testing.T) {
	tests := []struct {
		name      string
		newTicket model.Ticket
		want      model.Ticket
		wantErr   error
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
			setupSubTest(t, transitions)

			got, err := tc.CreateTicket(tt.newTicket)
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

func TestFixTicket(t *testing.T) {
	tests := []struct {
		name            string
		ticketID        string
		fixedWorkflow   []string
		transitions     map[string][]model.Transition
		wantTicket      model.Ticket
		wantTransitions []string
		wantErr         error
	}{
		{
			name:     "ResolveFromTodo",
			ticketID: "TEST-1",
			fixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			transitions: transitions,
			wantTicket: model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-1",
				Description: GenerateDescriptionText("Description TEST-1", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
				Labels:      []string{"Vulnerability"},
			},
			wantTransitions: []string{InProgress, Resolved},
			wantErr:         nil,
		},
		{
			name:     "FixInProgress",
			ticketID: "TEST-2",
			fixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			transitions: transitions,
			wantTicket: model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-2",
				Description: GenerateDescriptionText("Description TEST-2", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
				Labels:      []string{"Potential Vulnerability"},
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:          "ResolveFromToDoClassicWorkflow",
			ticketID:      "TEST-1",
			fixedWorkflow: []string{Resolved},
			transitions:   trResolveFromAnyStatus,
			wantTicket: model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-1",
				Description: GenerateDescriptionText("Description TEST-1", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
				Labels:      []string{"Vulnerability"},
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:          "ResolveFromInProgressClassicWorkflow",
			ticketID:      "TEST-2",
			fixedWorkflow: []string{Resolved},
			transitions:   trResolveFromAnyStatus,
			wantTicket: model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-2",
				Description: GenerateDescriptionText("Description TEST-2", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
				Labels:      []string{"Potential Vulnerability"},
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:     "FixTicketNotFound",
			ticketID: "NOTFOUND",
			fixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			transitions:     transitions,
			wantTicket:      model.Ticket{},
			wantTransitions: []string{},
			wantErr:         errors.New("ticket NOTFOUND not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, tt.transitions)

			got, err := tc.FixTicket(tt.ticketID, tt.fixedWorkflow)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.wantTicket)
			if diff != "" {
				t.Fatalf("ticket does not match expected one. diff: %s\n", diff)
			}
			diff = cmp.Diff(transitionsDone, tt.wantTransitions)
			if diff != "" {
				t.Fatalf("transitions does not match expected ones. diff: %s\n", diff)
			}
		})
	}
}

func TestWontFixTicket(t *testing.T) {
	tests := []struct {
		name              string
		ticketID          string
		wontfixedWorkflow []string
		transitions       map[string][]model.Transition
		wantTicket        model.Ticket
		wantTransitions   []string
		wantErr           error
	}{
		{
			name:     "WontFixFromToDo",
			ticketID: "TEST-1",
			wontfixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			transitions: transitions,
			wantTicket: model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-1",
				Description: GenerateDescriptionText("Description TEST-1", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
				Labels:      []string{"Vulnerability"},
			},
			wantTransitions: []string{InProgress, Resolved},
			wantErr:         nil,
		},
		{
			name:     "WontFixInProgress",
			ticketID: "TEST-2",
			wontfixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			transitions: transitions,
			wantTicket: model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-2",
				Description: GenerateDescriptionText("Description TEST-2", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
				Labels:      []string{"Potential Vulnerability"},
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:              "WontFixFromToDoClassicWorkflow",
			ticketID:          "TEST-1",
			wontfixedWorkflow: []string{Resolved},
			transitions:       trResolveFromAnyStatus,
			wantTicket: model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-1",
				Description: GenerateDescriptionText("Description TEST-1", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
				Labels:      []string{"Vulnerability"},
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:              "WontFixInProgressClassicWorkflow",
			ticketID:          "TEST-2",
			wontfixedWorkflow: []string{Resolved},
			transitions:       trResolveFromAnyStatus,
			wantTicket: model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				TeamID:      "ff9d5142-0eb3-494a-8626-a72b7182cdb2",
				FindingID:   "4c24526d-2651-4873-8024-b27361b69723",
				Summary:     "Summary TEST-2",
				Description: GenerateDescriptionText("Description TEST-2", "4c24526d-2651-4873-8024-b27361b69723", "ff9d5142-0eb3-494a-8626-a72b7182cdb2"),
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
				Labels:      []string{"Potential Vulnerability"},
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:     "WontFixInProgressTicket",
			ticketID: "NOTFOUND",
			wontfixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			transitions:     transitions,
			wantTicket:      model.Ticket{},
			wantTransitions: []string{},
			wantErr:         errors.New("ticket NOTFOUND not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, tt.transitions)

			got, err := tc.WontFixTicket(tt.ticketID, tt.wontfixedWorkflow, "Decision Taken")
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.wantTicket)
			if diff != "" {
				t.Fatalf("ticket does not match expected one. diff: %s\n", diff)
			}
			diff = cmp.Diff(transitionsDone, tt.wantTransitions)
			if diff != "" {
				t.Fatalf("transitions does not match expected ones. diff: %s\n", diff)
			}
		})
	}
}
