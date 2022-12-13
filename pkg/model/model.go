/*
Copyright 2002 Adevinta
*/

package model

// Ticket represents a vulnerability in a ticket tracker.
type Ticket struct {
	ID          string   `json:"id"`
	Key         string   `json:"key"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Project     string   `json:"project"`
	Status      string   `json:"status"`
	TicketType  string   `json:"ticket_type"`
	Resolution  string   `json:"resolution"`
	Labels      []string `json:"labels"`
}

// Transition represents a state change of a ticket.
type Transition struct {
	ID     string `json:"id"`
	ToName string `json:"name"`
}

// TrackerConfig represents the configuration of a ticket tracker server.
type TrackerConfig struct {
	Name string
	Url  string
	User string
	Pass string
	Kind string
}

// ProjectConfig represents the configuration of a team.
type ProjectConfig struct {
	Name                   string
	ServerName             string
	Project                string
	VulnerabilityIssueType string
	FixedWorkflow          []string
	WontFixWorkflow        []string
}
