/*
Copyright 2022 Adevinta
*/
package storage

import (
	"errors"
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	testutil "github.com/adevinta/vulcan-tracker/pkg/testutils"
	"github.com/google/go-cmp/cmp"
)

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}

func TestServersConf(t *testing.T) {

	tests := []struct {
		name    string
		input   TOMLStore
		want    []model.TrackerConfig
		wantErr error
	}{
		{
			name: "HappyPath",
			input: TOMLStore{
				servers: map[string]config.Server{
					"example1": {
						Url:   "http://localhost:8080",
						User:  "jira_user",
						Token: "jira_token",
						Kind:  "jira",
					},
				},
			},
			want: []model.TrackerConfig{
				{
					Name: "example1",
					Url:  "http://localhost:8080",
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

			got, err := tt.input.ServersConf()

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}
}

func TestProjectsConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   TOMLStore
		want    []model.ProjectConfig
		wantErr error
	}{
		{
			name: "HappyPath",
			input: TOMLStore{
				teams: map[string]config.Team{
					"example_team": {
						Server:                 "example1",
						Project:                "TEST",
						VulnerabilityIssueType: "Vulnerability",
						FixWorkflow:            []string{"ToDo", "In Progress", "Under Review", "Fixed"},
						WontFixWorkflow:        []string{"Won't Fix"},
					},
				},
			},
			want: []model.ProjectConfig{
				{
					Name:                   "example_team",
					ServerName:             "example1",
					Project:                "TEST",
					VulnerabilityIssueType: "Vulnerability",
					FixedWorkflow:          []string{"ToDo", "In Progress", "Under Review", "Fixed"},
					WontFixWorkflow:        []string{"Won't Fix"},
				},
			},

			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.input.ProjectsConfig()

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}

}

func TestProjectConfig(t *testing.T) {
	input := TOMLStore{
		teams: map[string]config.Team{
			"example_team": {
				Server:                 "example1",
				Project:                "TEST",
				VulnerabilityIssueType: "Vulnerability",
				FixWorkflow:            []string{"ToDo", "In Progress", "Under Review", "Fixed"},
				WontFixWorkflow:        []string{"Won't Fix"},
			},
		},
	}

	tests := []struct {
		name    string
		teamId  string
		want    *model.ProjectConfig
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamId: "example_team",
			want: &model.ProjectConfig{

				Name:                   "example_team",
				ServerName:             "example1",
				Project:                "TEST",
				VulnerabilityIssueType: "Vulnerability",
				FixedWorkflow:          []string{"ToDo", "In Progress", "Under Review", "Fixed"},
				WontFixWorkflow:        []string{"Won't Fix"},
			},
			wantErr: nil,
		},
		{
			name:    "Notfound",
			teamId:  "noteam",
			want:    nil,
			wantErr: errors.New("team noteam not found in toml configuration"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := input.ProjectConfig(tt.teamId)

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatal(err)
			}
			diff := cmp.Diff(got, tt.want)
			if diff != "" {
				t.Errorf("%v\n", diff)
			}
		})
	}

}
