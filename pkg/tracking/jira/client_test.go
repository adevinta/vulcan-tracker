package jira

import (
	"fmt"
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	gojira "github.com/andygrunwald/go-jira"
	"github.com/google/go-cmp/cmp"
)

type MockIssueService struct {
	tickets map[string]*gojira.Issue
}

// Get retrieves a tikcet by issueID
func (mis *MockIssueService) Get(issueID string, options *gojira.GetQueryOptions) (*gojira.Issue, *gojira.Response, error) {
	value, ok := mis.tickets[issueID]
	if ok {
		return value, nil, nil
	}
	return nil, nil, fmt.Errorf("Key %s not found. Status code: 404", issueID)
}

func (mis *MockIssueService) Create(issue *gojira.Issue) (*gojira.Issue, *gojira.Response, error) {
	return &gojira.Issue{}, nil, nil
}

func (mis *MockIssueService) GetTransitions(id string) ([]gojira.Transition, *gojira.Response, error) {
	return []gojira.Transition{}, nil, nil
}

func (mis *MockIssueService) DoTransition(ticketID, transitionID string) (*gojira.Response, error) {
	return nil, nil
}

func (mis *MockIssueService) DoTransitionWithPayload(ticketID, payload interface{}) (*gojira.Response, error) {
	return nil, nil
}

var jiraClient *Client

func setupSubTestClient(t *testing.T) {
	t.Log("setup sub test jira client")

	tickets := make(map[string]*gojira.Issue)
	tickets["TEST-1"] = &gojira.Issue{
		ID:  "1000",
		Key: "TEST-1",
		Fields: &gojira.IssueFields{
			Summary:     "Summary TEST-1",
			Description: "Description TEST-1",
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
	tickets["TEST-2"] = &gojira.Issue{
		ID:  "1001",
		Key: "TEST-2",
		Fields: &gojira.IssueFields{
			Summary:     "Summary TEST-2",
			Description: "Description TEST-2",
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

	goJiraClient := &gojira.Client{}
	jiraClient = &Client{
		c: goJiraClient,
		Issuer: &MockIssueService{
			tickets: tickets,
		},
	}

}

func TestClient_Get(t *testing.T) {
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
			setupSubTestClient(t)

			got, err := jiraClient.GetTicket(tt.ticketId)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

// TODO: Pending more testing
