/*
Copyright 2002 Adevinta
*/

package model

// Issue represents a vulnerability in an issue tracker.
type Issue struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Project     string `json:"project"`
	Status      string `json:"status"`
}

// Transition represents a state change of an issue.
type Transition struct {
	ID     string `json:"id"`
	ToName string `json:"name"`
}
