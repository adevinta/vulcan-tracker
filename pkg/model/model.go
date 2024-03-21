/*
Copyright 2002 Adevinta
*/

// Package model represents the data model for vulcan-tracker.
package model

// Ticket represents a vulnerability in a ticket tracker.
type Ticket struct {
	ID          string   `json:"id"`
	Key         string   `json:"key"`
	TeamID      string   `json:"team_id"`
	FindingID   string   `json:"finding_id"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Project     string   `json:"project"`
	Status      string   `json:"status"`
	TicketType  string   `json:"ticket_type"`
	Resolution  string   `json:"resolution"`
	Labels      []string `json:"labels"`
	URLTracker  string   `json:"url_tracker"`
}

// FindingTicket represents a ticket for a finding and a team.
type FindingTicket struct {
	ID         string `json:"id" db:"id"`
	FindingID  string `json:"finding_id" db:"finding_id"`
	TeamID     string `json:"team_id" db:"team_id"`
	URLTracker string `json:"url_tracker" db:"url_tracker"`
}

// TrackerConfig represents the configuration of a ticket tracker server.
type TrackerConfig struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	URL   string `db:"url"`
	Token string
}

// ProjectConfig represents the configuration of a team.
type ProjectConfig struct {
	ID                     string `db:"id"`
	Name                   string `db:"name"`
	TeamID                 string `db:"team_id"`
	TrackerConfigID        string `db:"tracker_configuration_id"`
	Project                string `db:"project"`
	VulnerabilityIssueType string `db:"issue_type"`
}
