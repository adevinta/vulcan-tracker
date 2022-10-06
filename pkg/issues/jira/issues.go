/*
Copyright 2022 Adevinta
*/

package jira

import (
	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// GetIssue retrieves an issue from Jira.
func (is IS) GetIssue(id string) (*model.Issue, error) {
	issue, err := is.Client.GetIssue(id)
	if err != nil {
		return nil, err
	}
	return issue, nil

}

// CreateIssue creates an issue in Jira.
func (is IS) CreateIssue(issue *model.Issue) (*model.Issue, error) {

	issue.Project = is.Project

	createdIssue, err := is.Client.CreateIssue(issue, is.VulnerabilityIssueType)
	if err != nil {
		return nil, err
	}

	return createdIssue, nil
}

// GetTransitions retrieves all the transitions that are posibles from the curruent state of an issue.
func (is IS) GetTransitions(id string) (*[]model.Transition, error) {
	transitions, err := is.Client.GetIssueTransitions(id)
	if err != nil {
		return nil, err
	}

	return transitions, nil
}

// FixIssue transits an issue until a "done" state.
func (is IS) FixIssue(id string) (*model.Issue, error) {

	for _, transitionName := range is.IssueWorkflow {

		transitions, err := is.Client.GetIssueTransitions(id)
		if err != nil {
			return nil, err
		}

		for _, transition := range *transitions {
			if transition.ToName == transitionName {
				err := is.Client.DoTransition(id, transition.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	issue, err := is.Client.GetIssue(id)
	if err != nil {
		return nil, err
	}

	return issue, nil
}
