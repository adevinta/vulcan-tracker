/*
Copyright 2022 Adevinta
*/
package storage

import (
	"fmt"

	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// TOMLStore holds toml configuration.
type TOMLStore struct {
	servers map[string]config.Server
	teams   map[string]config.Team
}

// New creates a new instance to red the configuration from a toml file.
func New(servers map[string]config.Server, teams map[string]config.Team) (*TOMLStore, error) {
	return &TOMLStore{
		servers: servers,
		teams:   teams,
	}, nil
}

// ServersConf retrieves a list of all server configurations declared in the toml file.
func (ts *TOMLStore) ServersConf() ([]model.TrackerConfig, error) {
	var trackerConfigs []model.TrackerConfig

	for serverName, server := range ts.servers {
		serverConf := model.TrackerConfig{
			Name: serverName,
			Url:  server.Url,
			User: server.User,
			Pass: server.Token,
			Kind: server.Kind,
		}
		trackerConfigs = append(trackerConfigs, serverConf)
	}

	return trackerConfigs, nil
}

// ProjectsConfig retrieves a list of all team configurations declared in the toml file.
func (ts *TOMLStore) ProjectsConfig() ([]model.ProjectConfig, error) {
	var projectConfigs []model.ProjectConfig

	for teamName, team := range ts.teams {
		teamConfig := model.ProjectConfig{
			Name:                   teamName,
			ServerName:             team.Server,
			Project:                team.Project,
			VulnerabilityIssueType: team.VulnerabilityIssueType,
			FixedWorkflow:          team.FixWorkflow,
			WontFixWorkflow:        team.WontFixWorkflow,
			AutoCreate:             team.AutoCreate,
		}
		projectConfigs = append(projectConfigs, teamConfig)
	}

	return projectConfigs, nil

}

// ProjectConfig retrieves the configuration for the team teamId.
func (ts *TOMLStore) ProjectConfig(teamId string) (*model.ProjectConfig, error) {
	team, ok := ts.teams[teamId]
	if !ok {
		return nil, fmt.Errorf("team %s not found in toml configuration", teamId)
	}

	projectConfig := &model.ProjectConfig{
		Name:                   teamId,
		ServerName:             team.Server,
		Project:                team.Project,
		VulnerabilityIssueType: team.VulnerabilityIssueType,
		FixedWorkflow:          team.FixWorkflow,
		WontFixWorkflow:        team.WontFixWorkflow,
		AutoCreate:             team.AutoCreate,
	}

	return projectConfig, nil
}
