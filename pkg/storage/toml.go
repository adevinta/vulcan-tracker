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
	servers  map[string]config.Server
	projects map[string]config.Project
}

// New creates a new instance to red the configuration from a toml file.
func New(servers map[string]config.Server, projects map[string]config.Project) (*TOMLStore, error) {
	return &TOMLStore{
		servers:  servers,
		projects: projects,
	}, nil
}

// ServersConf retrieves a list of all server configurations declared in the toml file.
func (ts *TOMLStore) ServersConf() ([]model.TrackerConfig, error) {
	var trackerConfigs []model.TrackerConfig

	for serverID, server := range ts.servers {
		serverConf := model.TrackerConfig{
			ID:   serverID,
			Name: server.Name,
			URL:  server.URL,
			User: server.User,
			Pass: server.Token,
			Kind: server.Kind,
		}
		trackerConfigs = append(trackerConfigs, serverConf)
	}

	return trackerConfigs, nil
}

// ServerConf retrieves a server configuration declared in the toml file.
func (ts *TOMLStore) ServerConf(serverID string) (model.TrackerConfig, error) {
	server, ok := ts.servers[serverID]
	if !ok {
		return model.TrackerConfig{}, fmt.Errorf("server %s not found in toml configuration", serverID)
	}

	serverConf := model.TrackerConfig{
		Name: server.Name,
		URL:  server.URL,
		User: server.User,
		Pass: server.Token,
		Kind: server.Kind,
	}

	return serverConf, nil
}

// ProjectConfigByTeamID retrieves the configuration for the team teamID.
func (ts *TOMLStore) ProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {

	for id, project := range ts.projects {
		if project.TeamID == teamID {
			projectConfig := model.ProjectConfig{
				ID:                     id,
				Name:                   project.Name,
				ServerID:               project.ServerID,
				Project:                project.Project,
				VulnerabilityIssueType: project.VulnerabilityIssueType,
				FixedWorkflow:          project.FixWorkflow,
				WontFixWorkflow:        project.WontFixWorkflow,
			}
			return projectConfig, nil
		}
	}

	return model.ProjectConfig{}, fmt.Errorf("project not found in toml configuration for the team %s", teamID)
}
