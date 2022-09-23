/*
Copyright 2022 Adevinta
*/

package issue_tracking

import (
	"github.com/adevinta/vulcan-jira-api/pkg/model"
)

// Filter holds query filtering information.
type Filter struct {
	Status          string
	Tag             string
	Tags            string
	Identifier      string
	Identifiers     string
	IdentifierMatch bool
	MinScore        float32
	MaxScore        float32
	MinDate         string
	MaxDate         string
	AtDate          string
	Page            int
	Size            int
	SortBy          SortBy
	IssueID         string
	TargetID        string
	SourceID        string
	Labels          string
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
