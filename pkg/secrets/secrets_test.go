/*
Copyright 2023 Adevinta
*/

package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"

	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/testutil"
)

type smMock struct {
	secretsmanageriface.SecretsManagerAPI
	storedCredentials map[string]Credentials
}

// Helper function to get a string pointer for input string.
func getStrPtr(str string) *string {
	return &str
}

func (sm *smMock) DescribeSecretWithContext(_ aws.Context, input *secretsmanager.DescribeSecretInput, _ ...request.Option) (*secretsmanager.DescribeSecretOutput, error) {
	versionID := getStrPtr("uuid")
	versionStages := []*string{getStrPtr("AWSCURRENT")}
	versionIdsToStages := make(map[string][]*string)
	versionIdsToStages[*versionID] = versionStages

	mockedDescribeResult := secretsmanager.DescribeSecretOutput{
		ARN:                getStrPtr("dummy-arn"),
		Name:               getStrPtr(*input.SecretId),
		Description:        getStrPtr("secret description description"),
		VersionIdsToStages: versionIdsToStages,
	}
	return &mockedDescribeResult, nil
}

func (sm *smMock) GetSecretValueWithContext(_ aws.Context, input *secretsmanager.GetSecretValueInput, _ ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	versionID := getStrPtr("uuid")
	versionStages := []*string{getStrPtr("AWSCURRENT")}
	versionIdsToStages := make(map[string][]*string)
	versionIdsToStages[*versionID] = versionStages

	createDate := time.Now().Add(-time.Hour * 12) // 12 hours ago.
	value, ok := sm.storedCredentials[*input.SecretId]
	if !ok {
		return nil, errors.New("credentials not found")
	}

	secretID := input.SecretId
	secretBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error marshalling the credentials: %w", err)
	}
	secretString := getStrPtr(string(secretBytes))

	mockedGetResult := secretsmanager.GetSecretValueOutput{
		ARN:           getStrPtr("dummy-arn"),
		CreatedDate:   &createDate,
		Name:          secretID,
		SecretString:  secretString,
		VersionId:     versionID,
		VersionStages: versionStages,
	}
	return &mockedGetResult, nil
}

type mockLogger struct {
	echo.Logger
}

func TestGetServerCredentials(t *testing.T) {
	storedCredentials := map[string]Credentials{
		"/path/to/key/078b038c-3da5-4694-ba6b-43ef2e52d75c": {
			Token: "token1",
		},
	}

	client := smMock{
		storedCredentials: storedCredentials,
	}
	sc, err := secretcache.New(func(c *secretcache.Cache) { c.Client = &client })
	if err != nil {
		t.Fatalf("Unexpected error - %s", err.Error())
	}

	secrets := AWSSecrets{
		Logger:      mockLogger{},
		secretCache: sc,
		config: config.AwsConfig{
			ServerCredentialsKey: "/path/to/key/",
		},
	}

	tests := []struct {
		name     string
		serverID string
		want     Credentials
		wantErr  error
	}{
		{
			name:     "HappyPath",
			serverID: "078b038c-3da5-4694-ba6b-43ef2e52d75c",
			want: Credentials{
				Token: "token1",
			},
			wantErr: nil,
		},
		{
			name:     "CredentialNotFound",
			serverID: "e63c8e02-de5e-4875-8bb6-14975d5cc717",
			want:     Credentials{},
			wantErr:  errors.New("unable to retrieve the credentials for the server"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := secrets.GetServerCredentials(tt.serverID)
			if !strings.Contains(testutil.ErrToStr(err), testutil.ErrToStr(tt.wantErr)) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Fatalf("credentials does not match expected one. diff: %s\n", diff)
			}
		})
	}
}
