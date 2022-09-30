package toml

import (
	"fmt"

	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// TC holds toml configuration.
type TomlStorage struct {
	servers map[string]config.Server
	teams   map[string]config.Team
}

// New creates a new instance to red the configuration from a toml file.
func New(servers map[string]config.Server, teams map[string]config.Team) (*TomlStorage, error) {
	return &TomlStorage{
		servers: servers,
		teams:   teams,
	}, nil

}

// ListTrackerServersConf retrieves a list of all server configurations declared in the toml file.
func (tc TomlStorage) ListTrackerServersConf() ([]model.TrackerServerConf, error) {

	var trackerServerConfs []model.TrackerServerConf

	for serverName, server := range tc.servers {
		serverConf := model.TrackerServerConf{
			Name: serverName,
			Url:  server.Url,
			User: server.User,
			Pass: server.Token,
			Kind: server.Kind,
		}
		trackerServerConfs = append(trackerServerConfs, serverConf)
	}

	return trackerServerConfs, nil
}

// ListTrackerConfigurations retrieves a list of all team configurations declared in the toml file.
func (tc TomlStorage) ListTrackerConfigurations() ([]model.TrackerConfiguration, error) {

	var trackerConfigurations []model.TrackerConfiguration

	for teamName, team := range tc.teams {
		teamConfig := model.TrackerConfiguration{
			Name:                   teamName,
			ServerName:             team.Server,
			Project:                team.Project,
			VulnerabilityIssueType: team.VulnerabilityIssueType,
			FixedWorkflow:          team.FixWorkflow,
			WontFixWorkflow:        team.WontFixWorkflow,
		}
		trackerConfigurations = append(trackerConfigurations, teamConfig)
	}

	return trackerConfigurations, nil

}

// GetTrackerConfiguration retrieves the configuration for the team teamId.
func (tc TomlStorage) GetTrackerConfiguration(teamId string) (*model.TrackerConfiguration, error) {

	team, ok := tc.teams[teamId]
	if !ok {
		return nil, fmt.Errorf("team %s not found in toml configuration", teamId)
	}

	teamConfig := &model.TrackerConfiguration{
		Name:                   teamId,
		ServerName:             team.Server,
		Project:                team.Project,
		VulnerabilityIssueType: team.VulnerabilityIssueType,
		FixedWorkflow:          team.FixWorkflow,
		WontFixWorkflow:        team.WontFixWorkflow,
	}

	return teamConfig, nil
}
