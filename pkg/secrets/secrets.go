/*
Copyright 2023 Adevinta
*/

package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/labstack/echo/v4"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const region = "eu-west-1"

type (
	// Secrets manage the credentials of ticket trackers servers.
	Secrets interface {
		GetServerCredentials(serverID string) (Credentials, error)
	}

	// AWSSecrets represents the specific AWS secret manager.
	AWSSecrets struct {
		Logger echo.Logger
	}

	// Credentials store the credentials of a secret.
	Credentials struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}
)

// GetServerCredentials return the user and password inside a Credentials type
// from AWS secret manager for a specific server.
func (s *AWSSecrets) GetServerCredentials(serverID string) (Credentials, error) {

	secretName := fmt.Sprintf("/vulcan/k8s/tracker/jira/%s", serverID)
	config, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(region))
	if err != nil {
		return Credentials{}, err
	}

	// Create Secrets Manager client.
	svc := secretsmanager.NewFromConfig(config)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified.
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		// For a list of exceptions thrown, see
		// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
		return Credentials{}, &vterrors.TrackingError{
			Msg:            fmt.Sprintf("Unable to retrieve the user and password for the server "),
			Err:            err,
			HTTPStatusCode: http.StatusUnauthorized,
		}
	}

	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString
	var credentials Credentials
	err = json.Unmarshal([]byte(secretString), &credentials)
	if err != nil {
		return Credentials{}, fmt.Errorf("error unmarshalling the credentials: %w", err)
	}

	return credentials, nil
}
