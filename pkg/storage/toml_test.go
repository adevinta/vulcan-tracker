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

func TestProjectConfigByTeamID(t *testing.T) {
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
		want    model.ProjectConfig
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "example_team",
			want: model.ProjectConfig{
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
			want:    model.ProjectConfig{},
			wantErr: errors.New("project not found in toml configuration for the team noteam"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tc := TOMLStore{
				projects: projects,
			}
			got, err := tc.FindProjectConfigByTeamID(tt.teamID)

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
