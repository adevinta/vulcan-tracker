package jira

import (
	"fmt"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	gojira "github.com/andygrunwald/go-jira"
	"github.com/trivago/tgo/tcontainer"
)

type Client struct {
	c *gojira.Client
}

// NewClient instanciates a client using go-jira library.
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
		c: gojiraClient,
	}, nil
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
func (cl *Client) GetTicket(id string) (*model.Ticket, error) {

	jiraIssue, _, err := cl.c.Issue.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return fromGoJiraToTicketModel(*jiraIssue), nil

}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

// CreateTicket creates a ticket in Jira.
func (cl *Client) CreateTicket(ticket *model.Ticket, issueType string) (*model.Ticket, error) {

	queryOptions := &gojira.GetQueryOptions{
		ProjectKeys: ticket.Project,
		Expand:      "projects.issuetypes.fields",
	}

	// Retrieve the metadata needed to create a ticket.
	createMetaInfo, _, err := cl.c.Issue.GetCreateMetaWithOptions(queryOptions)
	if err != nil {
		return nil, err
	}

	metaProject := createMetaInfo.GetProjectWithKey(ticket.Project)
	metaIssueType := metaProject.GetIssueTypeWithName(ticket.TicketType)
	mandatoryFields, err := metaIssueType.GetMandatoryFields()
	if err != nil {
		return nil, err
	}

	customfields := tcontainer.NewMarshalMap()

	for _, field := range mandatoryFields {
		// We already have the values for this fields.
		if contains([]string{"summary", "issuetype", "components", "project", "reporter"}, field) {
			continue
		}

		required, err := metaIssueType.Fields.Bool(field + "/required")
		if err != nil {
			return nil, err
		}
		hasDefaultValue, err := metaIssueType.Fields.Bool(field + "/hasDefaultValue")
		if err != nil {
			return nil, err
		}

		//We only have to manage the required fields without default value.
		if required && !hasDefaultValue {

			fieldType, exists := metaIssueType.Fields.Value(field + "/schema/type")
			if !exists {
				return nil, fmt.Errorf("error retrieving the type of the custom field %s", field)
			}
			switch fieldType {
			case "option":
				var firstOption interface{}
				firstOption, exists = metaIssueType.Fields.Value(field + "/allowedValues[0]value")
				if !exists {
					return nil, fmt.Errorf("error retrieving the first option of the custom field %s", field)
				}
				customfields[field] = map[string]string{"value": firstOption.(string)}
			}
		}

	}

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
			Unknowns: customfields,
		},
	}

	gojiraIssue, _, err := cl.c.Issue.Create(newTicket)
	if err != nil {
		return nil, err
	}

	createdTicket, _, err := cl.c.Issue.Get(gojiraIssue.ID, nil)
	if err != nil {
		return nil, err
	}
	return fromGoJiraToTicketModel(*createdTicket), nil
}

// GetTicketTransitions retrieves a list of all available transitions of a ticket.
func (cl *Client) GetTicketTransitions(id string) ([]model.Transition, error) {
	transitions, _, err := cl.c.Issue.GetTransitions(id)
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
func (cl *Client) DoTransition(id string, idTransition string) error {
	_, err := cl.c.Issue.DoTransition(id, idTransition)
	if err != nil {
		return err
	}
	return nil
}
