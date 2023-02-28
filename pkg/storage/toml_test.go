/*
Copyright 2022 Adevinta
*/
package storage

import (
	"errors"
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/testutils"
	"github.com/google/go-cmp/cmp"
)

func errToStr(err error) string {
	return testutils.ErrToStr(err)
}

func TestServersConf(t *testing.T) {

	tests := []struct {
		name    string
		servers map[string]config.Server
		want    []model.TrackerConfig
		wantErr error
	}{
		{
			name: "HappyPath",
			servers: map[string]config.Server{
				"example1_id": {
					Name:  "example1",
					URL:   "http://localhost:8080",
					User:  "jira_user",
					Token: "jira_token",
					Kind:  "jira",
				},
			},
			want: []model.TrackerConfig{
				{
					ID:   "example1_id",
					Name: "example1",
					URL:  "http://localhost:8080",
					User: "jira_user",
					Pass: "jira_token",
					Kind: "jira",
				},
			},

			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tc := TOMLStore{servers: tt.servers}
			got, err := tc.ServersConf()

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Fatalf("the servers do not match expected ones. diff: %v\n", diff)
			}
		})
	}
}

func TestProjectConfig(t *testing.T) {
	projects := map[string]config.Project{
		"example_project_id": {
			ServerID:               "example_server_id",
			Name:                   "example_team_name",
			TeamID:                 "example_team",
			Project:                "TEST",
			VulnerabilityIssueType: "Vulnerability",
			FixWorkflow:            []string{"ToDo", "In Progress", "Under Review", "Fixed"},
			WontFixWorkflow:        []string{"Won't Fix"},
		},
	}

	tests := []struct {
		name    string
		teamID  string
		want    *model.ProjectConfig
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "example_team",
			want: &model.ProjectConfig{
				ID:                     "example_project_id",
				Name:                   "example_team_name",
				ServerID:               "example_server_id",
				Project:                "TEST",
				VulnerabilityIssueType: "Vulnerability",
				FixedWorkflow:          []string{"ToDo", "In Progress", "Under Review", "Fixed"},
				WontFixWorkflow:        []string{"Won't Fix"},
			},
			wantErr: nil,
		},
		{
			name:    "Notfound",
			teamID:  "noteam",
			want:    nil,
			wantErr: errors.New("project not found in toml configuration for the team noteam"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tc := TOMLStore{
				projects: projects,
			}
			got, err := tc.ProjectConfigByTeamID(tt.teamID)

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Fatalf("the project does not match expected one. diff: %v\n", diff)
			}
		})
	}

}
