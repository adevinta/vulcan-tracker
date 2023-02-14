/*
Copyright 2022 Adevinta
*/
package tracking

import (
	"errors"
	"fmt"
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/secrets"
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	testutil "github.com/adevinta/vulcan-tracker/pkg/testutils"
	"github.com/adevinta/vulcan-tracker/pkg/tracking/jira"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
)

var ts *TS

type mockLogger struct {
	echo.Logger
}

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}

type mockTicketServerStorage struct {
	storage.TicketServerStorage
	servers  map[string]*model.TrackerConfig
	projects map[string]*model.ProjectConfig
}

func (mt *mockTicketServerStorage) FindServerConf(serverID string) (model.TrackerConfig, error) {
	value, ok := mt.servers[serverID]
	if !ok {
		return model.TrackerConfig{}, fmt.Errorf("server %s not found in the configuration", serverID)
	}
	return *value, nil
}

func (mt *mockTicketServerStorage) FindProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {
	for _, project := range mt.projects {
		if project.TeamID == teamID {
			return *project, nil
		}
	}

	return model.ProjectConfig{}, fmt.Errorf("project not found for the team %s", teamID)
}

type mockSecrets struct {
	credentials map[string]*secrets.Credentials
}

func (ms *mockSecrets) GetServerCredentials(serverID string) (secrets.Credentials, error) {
	value, ok := ms.credentials[serverID]
	if !ok {
		return secrets.Credentials{}, fmt.Errorf("credentials for server %s not found", serverID)
	}
	return *value, nil
}

func setupSubTest(t *testing.T) {
	t.Log("setup sub test")
	servers := make(map[string]*model.TrackerConfig)
	servers["JiraServerID"] = &model.TrackerConfig{
		ID:   "JiraServerID",
		Name: "JiraServer",
		URL:  "http://example.com",
		Kind: "jira",
	}
	servers["JiraServerIDNoCredentials"] = &model.TrackerConfig{
		ID:   "JiraServerIDNoCredentials",
		Name: "JiraServer",
		URL:  "http://example.com",
		Kind: "jira",
	}

	projects := make(map[string]*model.ProjectConfig)
	projects["projectID1"] = &model.ProjectConfig{
		ID:                     "projectID1",
		Name:                   "ProjectName1",
		TeamID:                 "ProjectTeamID1",
		ServerID:               "JiraServerID",
		Project:                "TEST-1",
		VulnerabilityIssueType: "Vulnerability",
		FixedWorkflow:          nil,
		WontFixWorkflow:        nil,
		AutoCreate:             false,
	}

	credentials := make(map[string]*secrets.Credentials)
	credentials["JiraServerID"] = &secrets.Credentials{
		User:     "user",
		Password: "password",
	}
	ts = &TS{
		ticketServerStorage: &mockTicketServerStorage{
			servers:  servers,
			projects: projects},
		secretsService: &mockSecrets{credentials: credentials},
		Logger:         &mockLogger{},
	}
}

func TestGenerateServerClients(t *testing.T) {
	logger := &mockLogger{}
	tests := []struct {
		name          string
		trackerConfig []model.TrackerConfig
		want          map[string]TicketTracker
		wantErr       error
	}{
		{
			name: "HappyPath",
			trackerConfig: []model.TrackerConfig{
				{
					ID:   "JiraServerID",
					Name: "JiraServer",
					URL:  "http://example.com",
					User: "user",
					Pass: "pass",
					Kind: "jira",
				},
			},
			want: map[string]TicketTracker{
				"JiraServerID": &jira.TC{
					Client: &jira.Client{},
					Logger: logger,
					URL:    "http://example.com",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := GenerateServerClients(tt.trackerConfig, logger)

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want, cmpopts.IgnoreUnexported(jira.TC{}),
				cmpopts.IgnoreUnexported(jira.Client{}),
				cmpopts.IgnoreInterfaces(struct{ jira.Issuer }{}),
			)
			if diff != "" {
				t.Fatalf("the generated servers do not match expected ones. diff: %v\n", diff)
			}
		})
	}
}

func TestServerConf(t *testing.T) {
	tests := []struct {
		name     string
		serverID string
		want     model.TrackerConfig
		wantErr  error
	}{
		{
			name:     "HappyPath",
			serverID: "JiraServerID",
			want: model.TrackerConfig{
				ID:   "JiraServerID",
				Name: "JiraServer",
				URL:  "http://example.com",
				User: "user",
				Pass: "password",
				Kind: "jira",
			},
			wantErr: nil,
		},
		{
			name:     "NoCredentials",
			serverID: "JiraServerIDNoCredentials",
			want:     model.TrackerConfig{},
			wantErr:  errors.New("credentials for server JiraServerIDNoCredentials not found"),
		},
		{
			name:     "NoServer",
			serverID: "NoServer",
			want:     model.TrackerConfig{},
			wantErr:  errors.New("server NoServer not found in the configuration"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t)
			got, err := ts.ServerConf(tt.serverID)

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want, cmpopts.IgnoreUnexported(jira.TC{}),
				cmpopts.IgnoreUnexported(jira.Client{}),
				cmpopts.IgnoreInterfaces(struct{ jira.Issuer }{}),
			)
			if diff != "" {
				t.Fatalf("the generated servers do not match expected ones. diff: %v\n", diff)
			}
		})
	}
}

func TestProjectConfigByTeamID(t *testing.T) {
	tests := []struct {
		name    string
		teamID  string
		want    model.ProjectConfig
		wantErr error
	}{
		{
			name:   "HappyPath",
			teamID: "ProjectTeamID1",
			want: model.ProjectConfig{
				ID:                     "projectID1",
				Name:                   "ProjectName1",
				TeamID:                 "ProjectTeamID1",
				ServerID:               "JiraServerID",
				Project:                "TEST-1",
				VulnerabilityIssueType: "Vulnerability",
				FixedWorkflow:          nil,
				WontFixWorkflow:        nil,
				AutoCreate:             false,
			},
			wantErr: nil,
		},
		{
			name:    "NoProject",
			teamID:  "NoProjectTeamID",
			want:    model.ProjectConfig{},
			wantErr: errors.New("project not found for the team NoProjectTeamID"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupSubTest(t)
			got, err := ts.ProjectConfigByTeamID(tt.teamID)

			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(got, tt.want, cmpopts.IgnoreUnexported(jira.TC{}),
				cmpopts.IgnoreUnexported(jira.Client{}),
				cmpopts.IgnoreInterfaces(struct{ jira.Issuer }{}),
			)
			if diff != "" {
				t.Fatalf("the generated project configurations do not match expected ones. diff: %v\n", diff)
			}
		})
	}
}
