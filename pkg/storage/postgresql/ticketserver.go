/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// FindServerConf retrieves a ticket tracker configuration from a postgres database.
func (db *PostgresStore) FindServerConf(serverID string) (model.TrackerConfig, error) {
	query := "SELECT * FROM tracker_configurations where id = $1"
	logQuery(db.Logger, "ServerConf", query, serverID)
	result := db.DB.QueryRow(query, serverID)

	var trackerConfig model.TrackerConfig
	err := result.Scan(&trackerConfig.ID, &trackerConfig.Name, &trackerConfig.URL)
	if err != nil {
		return model.TrackerConfig{}, err
	}

	return trackerConfig, nil
}

// FindProjectConfigByTeamID retrieves a project configuration for a specific team from a postgres database.
func (db *PostgresStore) FindProjectConfigByTeamID(teamID string) (model.ProjectConfig, error) {
	var projectConfig model.ProjectConfig

	query := "SELECT * FROM projects WHERE team_id = $1"
	logQuery(db.Logger, "GetFindingTicket", query, teamID)
	err := db.DB.Get(&projectConfig, query, teamID)
	if err != nil {
		return model.ProjectConfig{}, err
	}

	return projectConfig, nil
}
