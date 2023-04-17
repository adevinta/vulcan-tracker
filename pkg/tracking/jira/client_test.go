/*
Copyright 2022 Adevinta
*/
package jira

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	gojira "github.com/andygrunwald/go-jira"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type MockIssueService struct {
	Issuer
	tickets map[string]gojira.Issue
}

// Get retrieves a ticket by issueID.
func (mis *MockIssueService) Get(issueID string, _ *gojira.GetQueryOptions) (*gojira.Issue, *gojira.Response, error) {
	value, ok := mis.tickets[issueID]
	if ok {
		return &value, nil, nil
	}
	return nil, nil, fmt.Errorf("key %s not found. Status code: 404", issueID)
}

// Search tries to find a ticket.
func (mis *MockIssueService) Search(jql string, _ *gojira.SearchOptions) ([]gojira.Issue, *gojira.Response, error) {

	splitted := strings.Split(jql, " ")
	project := strings.Split(splitted[0], "=")[1]
	issueType := strings.Split(splitted[2], "=")[1]
	descriptionText := strings.Split(splitted[4], "~")[1]

	var ticketsFound []gojira.Issue

	for _, ticket := range mis.tickets {
		if ticket.Fields.Project.Key != project {
			continue
		}
		if ticket.Fields.Type.Name != issueType {
			continue
		}
		if strings.Contains(ticket.Fields.Description, descriptionText) {
			ticketsFound = append(ticketsFound, ticket)
		}

	}

	return ticketsFound, nil, nil
}

// Create creates a new ticket.
func (mis *MockIssueService) Create(issue *gojira.Issue) (*gojira.Issue, *gojira.Response, error) {

	if issue.Fields.Summary == "" {
		return nil, nil, fmt.Errorf("summary is mandatory. Status code: 400")
	}
	if issue.Fields.Type.Name == "" {
		return nil, nil, fmt.Errorf("issue type is mandatory. Status code: 400")
	}

	issue.Key = fmt.Sprintf("%s-%d", issue.Fields.Project.Key, len(mis.tickets)+1)
	issue.ID = fmt.Sprintf("%d", 1000+len(mis.tickets)+1)
	issue.Fields.Status = &gojira.Status{Name: ToDo}
	mis.tickets[issue.Key] = *issue
	return issue, nil, nil
}

var jiraClient *Client

func setupSubTestClient(t *testing.T) {
	t.Log("setup sub test jira client")

	tickets := make(map[string]gojira.Issue)
	tickets["TEST-1"] = gojira.Issue{
		ID:  "1000",
		Key: "TEST-1",
		Fields: &gojira.IssueFields{
			Summary:     "Summary TEST-1",
			Description: "Description TEST-1\nFindingID: FindingID-1",
			Project: gojira.Project{
				Key: "TEST",
			},
			Status: &gojira.Status{
				Name: ToDo,
			},
			Type: gojira.IssueType{
				Name: "Vulnerability",
			},
		},
	}
	tickets["TEST-2"] = gojira.Issue{
		ID:  "1001",
		Key: "TEST-2",
		Fields: &gojira.IssueFields{
			Summary:     "Summary TEST-2",
			Description: "Description TEST-2\nFindingID: FindingID-2",
			Project: gojira.Project{
				Key: "TEST",
			},
			Status: &gojira.Status{
				Name: InProgress,
			},
			Type: gojira.IssueType{
				Name: "Vulnerability",
			},
		},
	}

	jiraClient = &Client{
		Issuer: &MockIssueService{
			tickets: tickets,
		},
	}
}

func TestClient_Get(t *testing.T) {
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
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1\nFindingID: FindingID-1",
				Project:     "TEST",
				Status:      ToDo,
				TicketType:  "Vulnerability",
			},
			wantErr: nil,
		},
		{
			name:     "KeyNotFound",
			ticketID: "NOTFOUND",
			want:     model.Ticket{},
			wantErr:  errors.New("ticket NOTFOUND not found in Jira"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTestClient(t)

			got, err := jiraClient.GetTicket(tt.ticketID)
			if !isSameError(err, tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}

			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Fatalf("ticket does not match expected one. diff: %s\n", diff)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	tests := []struct {
		name      string
		newTicket model.Ticket
		want      model.Ticket
		wantErr   error
	}{
		{
			name: "HappyPath",
			newTicket: model.Ticket{
				Summary:     "Summary",
				Description: "Description",
				Project:     "TEST",
				TicketType:  "Vulnerability",
				Labels:      []string{"Potential Vulnerability"},
			},
			want: model.Ticket{
				Summary:     "Summary",
				Description: "Description",
				Project:     "TEST",
				Status:      ToDo,
				TicketType:  "Vulnerability",
				Labels:      []string{"Potential Vulnerability"},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTestClient(t)

			got, err := jiraClient.CreateTicket(tt.newTicket)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
			diff := cmp.Diff(got, tt.want, cmpopts.IgnoreFields(model.Ticket{}, "ID", "Key"))
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestClient_Find(t *testing.T) {
	tests := []struct {
		name                   string
		findingID              string
		projectKey             string
		vulnerabilityIssueType string
		want                   model.Ticket
		wantErr                error
	}{
		{
			name:                   "HappyPath",
			findingID:              "FindingID-1",
			projectKey:             "TEST",
			vulnerabilityIssueType: "Vulnerability",
			want: model.Ticket{
				ID:          "1000",
				Key:         "TEST-1",
				Summary:     "Summary TEST-1",
				Description: "Description TEST-1\nFindingID: FindingID-1",
				Project:     "TEST",
				Status:      ToDo,
				TicketType:  "Vulnerability",
			},
			wantErr: nil,
		},
		{
			name:                   "TicketNotFound",
			findingID:              "NOTFOUND",
			projectKey:             "TEST",
			vulnerabilityIssueType: "Vulnerability",
			want:                   model.Ticket{},
			wantErr:                nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTestClient(t)

			got, err := jiraClient.FindTicket(tt.projectKey, tt.vulnerabilityIssueType, tt.findingID)
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
