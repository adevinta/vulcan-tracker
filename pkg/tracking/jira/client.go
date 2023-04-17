/*
Copyright 2022 Adevinta
*/
package jira

import (
	"fmt"
	"net/http"
	"strings"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/adevinta/vulcan-tracker/pkg/model"

	gojira "github.com/andygrunwald/go-jira"
)

// Issuer manages all the operations related with the ticket tracker issues.
type Issuer interface {
	Get(issueID string, options *gojira.GetQueryOptions) (*gojira.Issue, *gojira.Response, error)
	Search(jql string, options *gojira.SearchOptions) ([]gojira.Issue, *gojira.Response, error)
	Create(issue *gojira.Issue) (*gojira.Issue, *gojira.Response, error)
}

// Client represents a specific Jira client.
type Client struct {
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
		Issuer: gojiraClient.Issue,
	}, nil
}

// fromGoJiraToTicketModel transforms a ticket returned by go-jira into a model.Ticket.
func fromGoJiraToTicketModel(jiraIssue gojira.Issue, ticket model.Ticket) model.Ticket {

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
	return ticket
}

// GetTicket retrieves a ticket from Jira.
// Return an empty ticket if not found.
func (cl *Client) GetTicket(id string) (model.Ticket, error) {
	jiraIssue, resp, err := cl.Issuer.Get(id, nil)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		if strings.Contains(err.Error(), "404") {
			return model.Ticket{}, &vterrors.TrackingError{
				Err:            fmt.Errorf("ticket %s not found in Jira: %w", id, err),
				HTTPStatusCode: http.StatusNotFound,
			}
		}
		return model.Ticket{}, err
	}
	var ticket model.Ticket
	ticket = fromGoJiraToTicketModel(*jiraIssue, ticket)
	return ticket, nil

}

// FindTicket search tickets and return the first one if it exists.
// The arguments needed to search a ticket are the project key, the issue
// type and a text that have to be present on the ticket description.
// Return a nil ticket if not found.
func (cl *Client) FindTicket(projectKey, vulnerabilityIssueType, text string) (model.Ticket, error) {

	jql := fmt.Sprintf("project=%s AND type=%s AND description~%s",
		projectKey, vulnerabilityIssueType, text)

	searchOptions := &gojira.SearchOptions{
		MaxResults: 1,
	}
	tickets, resp, err := cl.Issuer.Search(jql, searchOptions)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		return model.Ticket{}, err
	}
	if len(tickets) == 0 {
		return model.Ticket{}, nil
	}
	var ticket model.Ticket
	ticket = fromGoJiraToTicketModel(tickets[0], ticket)
	return ticket, nil
}

// CreateTicket creates a ticket in Jira.
func (cl *Client) CreateTicket(ticket model.Ticket) (model.Ticket, error) {
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
		return model.Ticket{}, err
	}

	createdTicket, resp, err := cl.Issuer.Get(gojiraIssue.Key, nil)
	if err != nil {
		err = gojira.NewJiraError(resp, err)
		return model.Ticket{}, err
	}

	ticket = fromGoJiraToTicketModel(*createdTicket, ticket)
	return ticket, nil
}
