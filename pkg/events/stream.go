/*
Copyright 2022 Adevinta
*/
// Package stream allows to interact with different stream-processing
// platforms.
package events

import (
	"context"
	"time"
)

// Message represents a message coming from a stream.
type Message struct {
	Key     []byte
	Value   []byte
	Headers []MetadataEntry
}

// MetadataEntry represents a metadata entry.
type MetadataEntry struct {
	Key   []byte
	Value []byte
}

// A Processor represents a stream message processor.
type Processor interface {
	Process(ctx context.Context, entity string, h MsgHandler) error
}

// A MsgHandler processes a message.
type MsgHandler func(msg Message) error

// Target represents the target
// scope for a check execution.
type Target struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
}

// TargetTeams represents a target along with its associated teams.
type TargetTeams struct {
	Target
	Teams []string `json:"teams"`
}

// Issue represents a security vulnerability.
type Issue struct {
	ID              string   `json:"id"`
	Summary         string   `json:"summary"`
	CWEID           uint32   `json:"cwe_id"`
	Description     string   `json:"description"`
	Recommendations []string `json:"recommendations"`
	ReferenceLinks  []string `json:"reference_links"`
}

// IssueLabels represents an issue along with its associated labels.
type IssueLabels struct {
	Issue
	Labels []string `json:"labels"`
}

// SourceFamily represents the set of sources with same
// name, component and target.
type SourceFamily struct {
	Name      string `json:"name" db:"name"`
	Component string `json:"component" db:"component"`
	Target    string `json:"-"`
}

// Source represents a source
// which reports vulnerabilities.
type Source struct {
	ID       string    `json:"id" db:"id"`
	Instance string    `json:"instance" db:"instance"`
	Options  string    `json:"options" db:"options"`
	Time     time.Time `json:"time" db:"time"`
	SourceFamily
}

// Finding represents the relationship between an Issue and a Target.
type Finding struct {
	ID               string  `json:"id"`
	AffectedResource string  `json:"affected_resource"`
	Score            float32 `json:"score"`
	Status           string  `json:"status"`
	Details          string  `json:"details"`
	ImpactDetails    string  `json:"impact_details"`
	Issue            Issue   `json:"issue"`

	// TotalExposure contains the time, in hours, a finding was open during its
	// entire lifetime. For instance: if the finding was open for one day, then
	// it was fixed, open again for onther day and finally fixed, the
	// TotalExposure would be 48. If the finding is OPEN the TotalExposure also
	// includes the period between the last time it was found and now.
	TotalExposure int64 `json:"total_exposure"`

	// Resources stores the resources of the finding, if any, in as a non typed json.
	Resources Resources `json:"resources"`
	//OpenFinding defines the fields that only open findings have defined.
	// *OpenFinding
}

// Resources defines the structure of a the resources of a finding.
type Resources []ResourceGroup

// ResourceGroup reprents a resource in a finding.
type ResourceGroup struct {
	Name       string              `json:"name"`
	Attributes []string            `json:"attributes"`
	Resources  []map[string]string `json:"resources"`
}

// FindingExpanded represents a finding expanding the associated target, issue
// and source data.
type FindingExpanded struct {
	Finding
	Issue           IssueLabels `json:"issue"`
	Target          TargetTeams `json:"target"`
	Source          Source      `json:"source"`
	Resources       Resources   `json:"resources"`
	TotalExposure   int64       `json:"total_exposure"`
	CurrentExposure int64       `json:"current_exposure,omitempty"`
}

// FindingNotification represents a notification associated with a finding state change.
type FindingNotification struct {
	FindingExpanded
	// TODO: Tag field is defined here in order for the FindingNotification struct to be
	// backward compatible with the "old" vulnerability notifications format so integrations
	// with external components that still use the old representation (e.g.: Hermes) are
	// maintained and can be isolated in the notify pkg. Once these integrations are modified
	// to use the new notification format we'll have to decide if we make this tag field
	// serializable into JSON and keep using the tag field as identifier, or we migrate to use
	// the current team identifier.
	Tag string `json:"-"`
}
