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

// NewTomlStore creates a new instance to red the configuration from a toml file.
func NewTomlStore(servers map[string]config.Server, projects map[string]config.Project) (*TOMLStore, error) {
	return &TOMLStore{
		servers:  servers,
		projects: projects,
	}, nil
}

// FindServerConf retrieves a server configuration declared in the toml file.
func (ts *TOMLStore) FindServerConf(serverID string) (model.TrackerConfig, error) {
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

// FindProjectConfigByTeamID retrieves the configuration for the team teamID.
func (ts *TOMLStore) FindProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {

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
