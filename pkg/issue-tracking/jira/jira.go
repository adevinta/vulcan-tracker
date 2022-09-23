/*
Copyright 2022 Adevinta
*/

package jira

import (
	issue_tracking "github.com/adevinta/vulcan-jira-api/pkg/issue-tracking"
	gojira "github.com/andygrunwald/go-jira"
	"github.com/labstack/echo/v4"
)

type (
	IS struct {
		Client                 issue_tracking.IssueTrackingClient
		VulnerabilityIssueType string
		Project                string
		IssueWorkflow          []string
		Logger                 echo.Logger
	}

	// ConnStr holds the Jira connection information.
	ConnStr struct {
		Url                    string   `toml:"url"`
		User                   string   `toml:"user"`
		Token                  string   `toml:"token"`
		VulnerabilityIssueType string   `toml:"vulnerability_issue_type"`
		Project                string   `toml:"project"`
		IssueWorkflow          []string `toml:"issue_work_flow"`
	}
)

// NewIS instantiates a new Jira connection.
func NewIS(cs ConnStr, logger echo.Logger) (*IS, error) {

	tp := gojira.BasicAuthTransport{
		Username: cs.User,
		Password: cs.Token,
	}

	gojiraClient, _ := gojira.NewClient(tp.Client(), cs.Url)

	jc := JiraClient{
		Client: *gojiraClient,
	}

	return &IS{
		Client:                 jc,
		VulnerabilityIssueType: cs.VulnerabilityIssueType,
		Project:                cs.Project,
		Logger:                 logger,
	}, nil
}
