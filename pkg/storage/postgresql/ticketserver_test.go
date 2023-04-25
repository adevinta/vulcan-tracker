/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"database/sql"
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/google/go-cmp/cmp"
	"github.com/labstack/gommon/log"
)

func TestFindServerConf(t *testing.T) {
	testStore, err := PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)

	}
	defer testStore.Close()
	tests := []struct {
		name            string
		TrackerConfigID string
		want            model.TrackerConfig
		wantErr         error
	}{
		{
			name:            "HappyPath",
			TrackerConfigID: "54018792-173f-4457-ab51-953b25d3a448",
			want: model.TrackerConfig{
				ID:   "54018792-173f-4457-ab51-953b25d3a448",
				Name: "Jira Server",
				URL:  "https://jira-server.com",
			},
			wantErr: nil,
		},
		{
			name:            "ProjectNotFound",
			TrackerConfigID: "50d3146c-9bdc-4002-b43b-1466bfd0a8b8",
			want:            model.TrackerConfig{},
			wantErr:         sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStore.FindServerConf(tt.TrackerConfigID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Fatalf("server config does not match expected one. diff: %s\n", diff)
			}
		})
	}
}

func TestFindProjectConfigByTeamID(t *testing.T) {
	testStore, err := PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)

	}
	defer testStore.Close()
	tests := []struct {
		name    string
		teamID  string
		want    model.ProjectConfig
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "7b54fd77-91f3-46d5-8b08-828655bf076c",
			want: model.ProjectConfig{
				ID:                     "611aa257-c575-4721-a5ac-57265e75d3b8",
				Name:                   "test project",
				TeamID:                 "7b54fd77-91f3-46d5-8b08-828655bf076c",
				TrackerConfigID:        "54018792-173f-4457-ab51-953b25d3a448",
				Project:                "TEST",
				VulnerabilityIssueType: "Vulnerability",
			},
			wantErr: nil,
		},
		{
			name:    "ProjectNotFound",
			teamID:  "e104227e-3e24-41ea-ac30-50fe84873b35",
			want:    model.ProjectConfig{},
			wantErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStore.FindProjectConfigByTeamID(tt.teamID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Fatalf("project config does not match expected one. diff: %s\n", diff)
			}
		})
	}
}
