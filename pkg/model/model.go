/*
Copyright 2002 Adevinta
*/

package model

import "github.com/lib/pq"

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

// Transition represents a state change of a ticket.
type Transition struct {
	ID     string `json:"id"`
	ToName string `json:"name"`
}

// TrackerConfig represents the configuration of a ticket tracker server.
type TrackerConfig struct {
	ID   string `db:"id"`
	Name string `db:"name"`
	URL  string `db:"url"`
	User string
	Pass string
	Kind string `db:"kind"`
}

// ProjectConfig represents the configuration of a team.
type ProjectConfig struct {
	ID                     string         `db:"id"`
	Name                   string         `db:"name"`
	TeamID                 string         `db:"team_id"`
	ServerID               string         `db:"ticket_tracker_servers_id"`
	Project                string         `db:"project"`
	VulnerabilityIssueType string         `db:"issue_type"`
	FixedWorkflow          pq.StringArray `db:"fix_workflow"`
	WontFixWorkflow        pq.StringArray `db:"wont_fix_workflow"`
	AutoCreate             bool           `db:"auto_create"`
}
