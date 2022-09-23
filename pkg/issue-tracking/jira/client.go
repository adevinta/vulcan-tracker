package jira

import (
	"github.com/adevinta/vulcan-jira-api/pkg/model"
	gojira "github.com/andygrunwald/go-jira"
)

type JiraClient struct {
	Client gojira.Client
}

// fromGoJiraToIssueModel transforms a issue returned by go-jira into a model.Issue.
func fromGoJiraToIssueModel(jiraIssue gojira.Issue) *model.Issue {

	return &model.Issue{
		ID:          jiraIssue.ID,
		Key:         jiraIssue.Key,
		Summary:     jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
		Project:     jiraIssue.Fields.Project.Key,
		Status:      jiraIssue.Fields.Status.Name,
	}
}

// fromGoJiraToTransitionModel transforms a transition returned by go-jira into a model.Transition.
func fromGoJiraToTransitionModel(jiraTransition gojira.Transition) *model.Transition {

	return &model.Transition{
		ID:     jiraTransition.ID,
		ToName: jiraTransition.To.Name,
	}
}

// GetIssue retrieves an issue from Jira.
func (jc JiraClient) GetIssue(id string) (*model.Issue, error) {

	issue, _, err := jc.Client.Issue.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return fromGoJiraToIssueModel(*issue), nil

}

// CreateIssue creates an issue in Jira.
func (jc JiraClient) CreateIssue(issue *model.Issue, issueType string) (*model.Issue, error) {
	newIssue := &gojira.Issue{
		Fields: &gojira.IssueFields{
			Description: issue.Description,
			Summary:     issue.Summary,
			Type: gojira.IssueType{
				Name: issueType,
			},
			Project: gojira.Project{
				Key: issue.Project,
			},
		},
	}

	gojiraIssue, _, err := jc.Client.Issue.Create(newIssue)
	if err != nil {
		return nil, err
	}

	createdIssue, _, err := jc.Client.Issue.Get(gojiraIssue.ID, nil)
	if err != nil {
		return nil, err
	}
	return fromGoJiraToIssueModel(*createdIssue), nil
}

// GetIssueTransitions retrieves a list of all available transitions of an issue.
func (js JiraClient) GetIssueTransitions(id string) (*[]model.Transition, error) {
	transitions, _, err := js.Client.Issue.GetTransitions(id)
	if err != nil {
		return nil, err
	}

	var result []model.Transition

	for _, transition := range transitions {
		transformedTransition := fromGoJiraToTransitionModel(transition)
		result = append(result, *transformedTransition)

	}
	return &result, nil
}

// DoTransition changes the state of an issue to one of the available ones.
func (js JiraClient) DoTransition(id string, idTransition string) error {
	_, err := js.Client.Issue.DoTransition(id, idTransition)
	if err != nil {
		return err
	}
	return nil
}
