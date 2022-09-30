/*
Copyright 2002 Adevinta
*/

package model

// Ticket represents a vulnerability in a ticket tracker.
type Ticket struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Project     string `json:"project"`
	Status      string `json:"status"`
	TicketType  string `json:"tickettype"`
}

// Transition represents a state change of a ticket.
type Transition struct {
	ID     string `json:"id"`
	ToName string `json:"name"`
}

// TrackerServerConf represents the configuration of a ticket tracker server.
type TrackerServerConf struct {
	Name string
	Url  string
	User string
	Pass string
	Kind string
}

// TrackerConfiguration represents the configuration of a team.
type TrackerConfiguration struct {
	Name                   string
	ServerName             string
	Project                string
	VulnerabilityIssueType string
	FixedWorkflow          []string
	WontFixWorkflow        []string
}
