package jira

import (
	"fmt"
	"testing"

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
	tickets     map[string]*model.Ticket
	transitions map[string][]model.Transition
}

type mockLogger struct {
	echo.Logger
}

func (l *mockLogger) Errorf(format string, args ...interface{}) {
	// Do nothing logger
}

func (mj MockJiraClient) GetTicket(id string) (*model.Ticket, error) {
	value, ok := mj.tickets[id]
	if ok {
		return value, nil
	}
	return &model.Ticket{}, nil
}
func (mj MockJiraClient) CreateTicket(ticket *model.Ticket) (*model.Ticket, error) {
	ticket.Key = fmt.Sprintf("%s-%d", ticket.Project, len(mj.tickets)+1)
	ticket.ID = fmt.Sprintf("%d", 1000+len(mj.tickets)+1)
	ticket.Status = ToDo

	mj.tickets[ticket.Key] = ticket

	return ticket, nil
}

func (mj MockJiraClient) GetTicketTransitions(id string) ([]model.Transition, error) {
	ticket, err := mj.GetTicket(id)
	if err != nil {
		return nil, err
	}
	transitions, ok := mj.transitions[ticket.Status]
	if ok {
		return transitions, nil
	}
	return nil, fmt.Errorf("Transitions for %s not found.", id)

}
func (mj MockJiraClient) DoTransition(id, idTransition string) error {
	ticket, err := mj.GetTicket(id)
	if err != nil {
		return err
	}
	transitions, ok := mj.transitions[ticket.Status]
	if !ok {
		return fmt.Errorf("Transitions for %s not found.", id)
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
func (mj MockJiraClient) DoTransitionWithResolution(id, idTransition, resolution string) error {
	ticket, err := mj.GetTicket(id)
	if err != nil {
		return err
	}
	transitions, ok := mj.transitions[ticket.Status]
	if !ok {
		return fmt.Errorf("Transitions for %s not found.", id)
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
	transitions     map[string][]model.Transition = map[string][]model.Transition{
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
	trResolveFromAnyStatus map[string][]model.Transition = map[string][]model.Transition{
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
		Summary:     "Summary TEST-1",
		Description: "Description TEST-1",
		Project:     "TEST",
		Status:      ToDo,
		TicketType:  "Vulnerability",
	}
	tickets["TEST-2"] = &model.Ticket{
		ID:          "1001",
		Key:         "TEST-2",
		Summary:     "Summary TEST-2",
		Description: "Description TEST-2",
		Project:     "TEST",
		Status:      InProgress,
		TicketType:  "Vulnerability",
	}

	jiraClient := MockJiraClient{
		tickets:     tickets,
		transitions: transitions,
	}
	tc = &TC{
		Client: jiraClient,
		Logger: &mockLogger{},
	}
	transitionsDone = []string{}

}

func TestGetTicket(t *testing.T) {
	tests := []struct {
		name     string
		ticketId string
		want     *model.Ticket
		wantErr  error
	}{
		{
			name:     "HappyPath",
			ticketId: "TEST-1",
			want: &model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
				Project:     "TEST",
				Status:      ToDo,
				TicketType:  "Vulnerability",
			},
			wantErr: nil,
		},
		{
			name:     "KeyNotFound",
			ticketId: "NOTFOUND",
			want:     &model.Ticket{},
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, transitions)

			got, err := tc.GetTicket(tt.ticketId)
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
		ticketId string
		want     []model.Transition
		wantErr  error
	}{
		{
			name:     "HappyPath",
			ticketId: "TEST-1",
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

			got, err := tc.GetTransitions(tt.ticketId)
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
		want      *model.Ticket
		wantErr   error
	}{
		{
			name: "HappyPath",
			newTicket: model.Ticket{
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
				Project:     "TEST",
				TicketType:  "Vulnerability",
			},
			want: &model.Ticket{
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
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

			got, err := tc.CreateTicket(&tt.newTicket)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(model.Ticket{}, "ID", "Key"))
			if diff != "" {
				t.Fatalf("ticket does not match expected one. diff: %s\n", diff)
			}
		})
	}

}

func TestFixTicket(t *testing.T) {
	tests := []struct {
		name            string
		ticketId        string
		fixedWorkflow   []string
		wantTicket      *model.Ticket
		wantTransitions []string
		wantErr         error
	}{
		{
			name:     "HappyPath",
			ticketId: "TEST-1",
			fixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			wantTicket: &model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
			},
			wantTransitions: []string{InProgress, Resolved},
			wantErr:         nil,
		},
		{
			name:     "FixInProgressTicket",
			ticketId: "TEST-2",
			fixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			wantTicket: &model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				Summary:     "Summary TEST-2",
				Description: "Description TEST-2",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:     "FixTicketNotFound",
			ticketId: "NOTFOUND",
			fixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			wantTicket:      &model.Ticket{},
			wantTransitions: []string{},
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, transitions)

			got, err := tc.FixTicket(tt.ticketId, tt.fixedWorkflow)
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

func TestFixResolveFromAnyCase(t *testing.T) {
	tests := []struct {
		name            string
		ticketId        string
		fixedWorkflow   []string
		wantTicket      *model.Ticket
		wantTransitions []string
		wantErr         error
	}{
		{
			name:          "ResolveFromToDo",
			ticketId:      "TEST-1",
			fixedWorkflow: []string{Resolved},
			wantTicket: &model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:          "ResolveFromInProgress",
			ticketId:      "TEST-2",
			fixedWorkflow: []string{Resolved},
			wantTicket: &model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				Summary:     "Summary TEST-2",
				Description: "Description TEST-2",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Done",
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, trResolveFromAnyStatus)

			got, err := tc.FixTicket(tt.ticketId, tt.fixedWorkflow)
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
		ticketId          string
		wontfixedWorkflow []string
		wantTicket        *model.Ticket
		wantTransitions   []string
		wantErr           error
	}{
		{
			name:     "HappyPath",
			ticketId: "TEST-1",
			wontfixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			wantTicket: &model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
			},
			wantTransitions: []string{InProgress, Resolved},
			wantErr:         nil,
		},
		{
			name:     "WontFixInProgressTicket",
			ticketId: "TEST-2",
			wontfixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			wantTicket: &model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				Summary:     "Summary TEST-2",
				Description: "Description TEST-2",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:     "WontFixInProgressTicket",
			ticketId: "NOTFOUND",
			wontfixedWorkflow: []string{
				ToDo, InProgress, Resolved,
			},
			wantTicket:      &model.Ticket{},
			wantTransitions: []string{},
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, transitions)

			got, err := tc.WontFixTicket(tt.ticketId, tt.wontfixedWorkflow, "Decision Taken")
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

func TestWontFixResolveFromAnyCase(t *testing.T) {
	tests := []struct {
		name              string
		ticketId          string
		wontfixedWorkflow []string
		wantTicket        *model.Ticket
		wantTransitions   []string
		wantErr           error
	}{
		{
			name:              "WontFixFromToDo",
			ticketId:          "TEST-1",
			wontfixedWorkflow: []string{Resolved},
			wantTicket: &model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:              "WontFixInProgressTicket",
			ticketId:          "TEST-2",
			wontfixedWorkflow: []string{Resolved},
			wantTicket: &model.Ticket{
				ID:          "1001",
				Key:         "TEST-2",
				Summary:     "Summary TEST-2",
				Description: "Description TEST-2",
				Project:     "TEST",
				Status:      Resolved,
				TicketType:  "Vulnerability",
				Resolution:  "Decision Taken",
			},
			wantTransitions: []string{Resolved},
			wantErr:         nil,
		},
		{
			name:              "WontFixInProgressTicket",
			ticketId:          "NOTFOUND",
			wontfixedWorkflow: []string{Resolved},
			wantTicket:        &model.Ticket{},
			wantTransitions:   []string{},
			wantErr:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t, trResolveFromAnyStatus)

			got, err := tc.WontFixTicket(tt.ticketId, tt.wontfixedWorkflow, "Decision Taken")
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
