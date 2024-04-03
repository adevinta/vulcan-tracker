/*
Copyright 2023 Adevinta
*/

// Package secrets manages the storage of credentials for ticket trackers.
package secrets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"github.com/labstack/echo/v4"

	"github.com/adevinta/vulcan-tracker/pkg/config"
	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
)

// Secrets manage the credentials of ticket trackers servers.
type Secrets interface {
	GetServerCredentials(serverID string) (Credentials, error)
}

// AWSSecrets represents the specific AWS secret manager.
type AWSSecrets struct {
	Logger      echo.Logger
	secretCache *secretcache.Cache
	config      config.AwsConfig
}

// Credentials store the credentials of a secret.
type Credentials struct {
	Token string `json:"token"`
}

// NewAWSSecretManager instantiates a manager for AWS Secret Manager.
func NewAWSSecretManager(config config.AwsConfig, logger echo.Logger) (*AWSSecrets, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	awsCfg := aws.NewConfig()
	awsCfg = awsCfg.WithRegion(config.Region)
	if config.Endpoint != "" {
		awsCfg = awsCfg.WithEndpoint(config.Endpoint)
	}
	// Create Secrets Manager client.
	client := secretsmanager.New(sess, awsCfg)
	sc, err := secretcache.New(func(c *secretcache.Cache) { c.Client = client })

	return &AWSSecrets{
		Logger:      logger,
		secretCache: sc,
		config:      config,
	}, err
}

// GetServerCredentials return the Jira credentials inside a Credentials type
// from AWS secret manager for a specific server.
func (s *AWSSecrets) GetServerCredentials(serverID string) (Credentials, error) {
	secretName := filepath.Join(s.config.ServerCredentialsKey, serverID)
	result, err := s.secretCache.GetSecretString(secretName)
	if err != nil {
		return Credentials{}, &vterrors.TrackingError{
			Err:            fmt.Errorf("unable to retrieve the credentials for the server: %w", err),
			HTTPStatusCode: http.StatusUnauthorized,
		}
	}

	// Decrypts secret using the associated KMS key.
	var credentials Credentials
	err = json.Unmarshal([]byte(result), &credentials)
	if err != nil {
		return Credentials{}, fmt.Errorf("error unmarshalling the credentials: %w", err)
	}

	return credentials, nil
}
