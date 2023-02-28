/*
Copyright 2022 Adevinta
*/
package tracking

import (
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	testutil "github.com/adevinta/vulcan-tracker/pkg/testutils"
	"github.com/adevinta/vulcan-tracker/pkg/tracking/jira"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/echo/v4"
)

type mockLogger struct {
	echo.Logger
}

func errToStr(err error) string {
	return testutil.ErrToStr(err)
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
					Name: "JiraServer",
					URL:  "http://example.com",
					User: "user",
					Pass: "pass",
					Kind: "jira",
				},
			},
			want: map[string]TicketTracker{
				"JiraServer": &jira.TC{
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
