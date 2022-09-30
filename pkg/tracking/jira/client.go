package jira

import (
	"github.com/adevinta/vulcan-tracker/pkg/model"
	gojira "github.com/andygrunwald/go-jira"
)

type JiraClient struct {
	Client gojira.Client
}

// NewClient instanciates a client using go-jira library.
func NewClient(url, user, token string) (*gojira.Client, error) {
	tp := gojira.BasicAuthTransport{
		Username: user,
		Password: token,
	}
	gojiraClient, err := gojira.NewClient(tp.Client(), url)
	if err != nil {
		return nil, err
	}
	return gojiraClient, nil
}

// fromGoJiraToTicketModel transforms a ticket returned by go-jira into a model.Ticket.
func fromGoJiraToTicketModel(jiraIssue gojira.Issue) *model.Ticket {

	return &model.Ticket{
		ID:          jiraIssue.ID,
		Key:         jiraIssue.Key,
		Summary:     jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
		Project:     jiraIssue.Fields.Project.Key,
		Status:      jiraIssue.Fields.Status.Name,
		TicketType:  jiraIssue.Fields.Type.Name,
	}
}

// fromGoJiraToTransitionModel transforms a transition returned by go-jira into a model.Transition.
func fromGoJiraToTransitionModel(jiraTransition gojira.Transition) *model.Transition {

	return &model.Transition{
		ID:     jiraTransition.ID,
		ToName: jiraTransition.To.Name,
	}
}

// GetTicket retrieves a ticket from Jira.
func (jc JiraClient) GetTicket(id string) (*model.Ticket, error) {

	jiraIssue, _, err := jc.Client.Issue.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return fromGoJiraToTicketModel(*jiraIssue), nil

}

// CreateTicket creates a ticket in Jira.
func (jc JiraClient) CreateTicket(ticket *model.Ticket, issueType string) (*model.Ticket, error) {
	newTicket := &gojira.Issue{
		Fields: &gojira.IssueFields{
			Description: ticket.Description,
			Summary:     ticket.Summary,
			Type: gojira.IssueType{
				Name: issueType,
			},
			Project: gojira.Project{
				Key: ticket.Project,
			},
		},
	}

	gojiraIssue, _, err := jc.Client.Issue.Create(newTicket)
	if err != nil {
		return nil, err
	}

	createdTicket, _, err := jc.Client.Issue.Get(gojiraIssue.ID, nil)
	if err != nil {
		return nil, err
	}
	return fromGoJiraToTicketModel(*createdTicket), nil
}

// GetTicketTransitions retrieves a list of all available transitions of a ticket.
func (jc JiraClient) GetTicketTransitions(id string) ([]model.Transition, error) {
	transitions, _, err := jc.Client.Issue.GetTransitions(id)
	if err != nil {
		return nil, err
	}

	var result []model.Transition

	for _, transition := range transitions {
		transformedTransition := fromGoJiraToTransitionModel(transition)
		result = append(result, *transformedTransition)

	}
	return result, nil
}

// DoTransition changes the state of an issue to one of the available ones.
func (jc JiraClient) DoTransition(id string, idTransition string) error {
	_, err := jc.Client.Issue.DoTransition(id, idTransition)
	if err != nil {
		return err
	}
	return nil
}
