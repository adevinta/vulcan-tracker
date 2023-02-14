/*
Copyright 2022 Adevinta
*/
package jira

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"

	gojira "github.com/andygrunwald/go-jira"
)

type Issuer interface {
	Get(issueID string, options *gojira.GetQueryOptions) (*gojira.Issue, *gojira.Response, error)
	Search(jql string, options *gojira.SearchOptions) ([]gojira.Issue, *gojira.Response, error)
	Create(issue *gojira.Issue) (*gojira.Issue, *gojira.Response, error)
	GetTransitions(id string) ([]gojira.Transition, *gojira.Response, error)
	DoTransition(ticketID, transitionID string) (*gojira.Response, error)
	DoTransitionWithPayload(ticketID, payload interface{}) (*gojira.Response, error)
}

type Client struct {
	c *gojira.Client
	Issuer
}

// NewClient instantiates a client using go-jira library.
func NewClient(url, user, token string) (*Client, error) {
	tp := gojira.BasicAuthTransport{
		Username: user,
		Password: token,
	}
	gojiraClient, err := gojira.NewClient(tp.Client(), url)
	if err != nil {
		return nil, err
	}
	return &Client{
		c:      gojiraClient,
		Issuer: gojiraClient.Issue,
	}, nil
}

// fromGoJiraToTicketModel transforms a ticket returned by go-jira into a model.Ticket.
func fromGoJiraToTicketModel(jiraIssue gojira.Issue, ticket *model.Ticket) {

	ticket.ID = jiraIssue.ID
	ticket.Key = jiraIssue.Key
	ticket.Summary = jiraIssue.Fields.Summary
	ticket.Description = jiraIssue.Fields.Description
	ticket.Project = jiraIssue.Fields.Project.Key
	ticket.Status = jiraIssue.Fields.Status.Name
	ticket.TicketType = jiraIssue.Fields.Type.Name
	ticket.Resolution = ""
	ticket.Labels = jiraIssue.Fields.Labels

	if jiraIssue.Fields.Resolution != nil {
		ticket.Resolution = jiraIssue.Fields.Resolution.Name
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
// Return an empty ticket if not found.
func (cl *Client) GetTicket(id string) (*model.Ticket, error) {
	jiraIssue, resp, err := cl.Issuer.Get(id, nil)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		if strings.Contains(err.Error(), "404") {
			return nil, &errors.TrackingError{
				Err:            err,
				Msg:            fmt.Sprintf("ticket %s not found in Jira", id),
				HttpStatusCode: http.StatusNotFound,
			}
		}
		return nil, err
	}
	var ticket model.Ticket
	fromGoJiraToTicketModel(*jiraIssue, &ticket)
	return &ticket, nil

}

// FindTicket search tickets and return the first one if it exists.
// The arguments needed to search a ticket are the project key, the issue
// type and a text that have to be present on the ticket description.
// Return a nil ticket if not found.
func (cl *Client) FindTicket(projectKey, vulnerabilityIssueType, text string) (*model.Ticket, error) {

	jql := fmt.Sprintf("project=%s AND type=%s AND description~%s",
		projectKey, vulnerabilityIssueType, text)

	searchOptions := &gojira.SearchOptions{
		MaxResults: 1,
	}
	tickets, resp, err := cl.Issuer.Search(jql, searchOptions)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		return nil, err
	}
	if len(tickets) == 0 {
		return nil, nil
	}
	var ticket model.Ticket
	fromGoJiraToTicketModel(tickets[0], &ticket)
	return &ticket, nil
}

// CreateTicket creates a ticket in Jira.
func (cl *Client) CreateTicket(ticket *model.Ticket) (*model.Ticket, error) {
	newTicket := &gojira.Issue{
		Fields: &gojira.IssueFields{
			Description: ticket.Description,
			Summary:     ticket.Summary,
			Type: gojira.IssueType{
				Name: ticket.TicketType,
			},
			Project: gojira.Project{
				Key: ticket.Project,
			},
			Labels: ticket.Labels,
		},
	}

	gojiraIssue, resp, err := cl.Issuer.Create(newTicket)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		return nil, err
	}

	createdTicket, resp, err := cl.Issuer.Get(gojiraIssue.Key, nil)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		return nil, err
	}

	fromGoJiraToTicketModel(*createdTicket, ticket)
	return ticket, nil
}

// GetTicketTransitions retrieves a list of all available transitions of a ticket.
func (cl *Client) GetTicketTransitions(id string) ([]model.Transition, error) {
	transitions, resp, err := cl.Issuer.GetTransitions(id)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
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
func (cl *Client) DoTransition(id, idTransition string) error {
	resp, err := cl.Issuer.DoTransition(id, idTransition)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		return err
	}
	return nil
}

// DoTransitionWithResolution changes the state of an issue to a resolved one and set the resolution field.
func (cl *Client) DoTransitionWithResolution(id, idTransition, resolution string) error {
	customPayload := map[string]interface{}{
		"transition": gojira.TransitionPayload{
			ID: idTransition,
		},
		"fields": gojira.TransitionPayloadFields{
			Resolution: &gojira.Resolution{
				Name: resolution,
			},
		},
	}

	resp, err := cl.Issuer.DoTransitionWithPayload(id, customPayload)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		return err
	}
	return nil
}
