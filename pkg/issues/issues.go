/*
Copyright 2022 Adevinta
*/

package issues

import (
	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// Filter holds query filtering information.
type Filter struct {
	// TODO: Not specified yet
}

// SortBy holds information for the
// query sorting criteria.
type SortBy struct {
	Field string
	Order string
}

// Pagination holds database pagination information.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// IssueTracking defines the API for high level querying data from issue tracker.
type IssueTracking interface {
	GetIssue(id string) (*model.Issue, error)
	CreateIssue(issue *model.Issue) (*model.Issue, error)
	GetTransitions(id string) (*[]model.Transition, error)
	FixIssue(id string) (*model.Issue, error)
}

// IssueTrackingClient defines the API of the adaptar for a third-party client.
type IssueTrackingClient interface {
	GetIssue(id string) (*model.Issue, error)
	CreateIssue(issue *model.Issue, issueType string) (*model.Issue, error)
	GetIssueTransitions(id string) (*[]model.Transition, error)
	DoTransition(id string, idTransition string) error
}
