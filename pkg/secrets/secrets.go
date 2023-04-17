/*
Copyright 2023 Adevinta
*/

package secrets

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adevinta/vulcan-tracker/pkg/config"
	vterrors "github.com/adevinta/vulcan-tracker/pkg/errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/labstack/echo/v4"

	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

// Secrets manage the credentials of ticket trackers servers.
type Secrets interface {
	GetServerCredentials(serverID string) (Credentials, error)
}

// AWSSecrets represents the specific AWS secret manager.
type AWSSecrets struct {
	Logger      echo.Logger
	secretCache *secretcache.Cache
	c           secretsmanageriface.SecretsManagerAPI
	config      config.AwsConfig
}

// Credentials store the credentials of a secret.
type Credentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// NewAWSSecretManager instantiates a manager for AWS Secret Manager.
func NewAWSSecretManager(config config.AwsConfig, logger echo.Logger) (*AWSSecrets, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	awsCfg := aws.NewConfig()
	awsCfg = awsCfg.WithRegion(config.Region)

	// Create Secrets Manager client.
	client := secretsmanager.New(sess, awsCfg)
	sc, err := secretcache.New(func(c *secretcache.Cache) { c.Client = client })

	return &AWSSecrets{
		Logger:      logger,
		secretCache: sc,
		c:           client,
		config:      config,
	}, err
}

// GetServerCredentials return the user and password inside a Credentials type
// from AWS secret manager for a specific server.
func (s *AWSSecrets) GetServerCredentials(serverID string) (Credentials, error) {
	secretName := fmt.Sprintf("%s%s", s.config.ServerCredentialsKey, serverID)

	result, err := s.secretCache.GetSecretString(secretName)
	if err != nil {
		return Credentials{}, &vterrors.TrackingError{
			Err:            fmt.Errorf("unable to retrieve the user and password for the server: %w", err),
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
