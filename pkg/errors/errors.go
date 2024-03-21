/*
Copyright 2022 Adevinta
*/

// Package errors manage the errors for vulcan-tracker.
package errors

// TrackingError represents a custom error for this project.
type TrackingError struct {
	Err            error
	HTTPStatusCode int
}

// Error returns an error as string.
func (te *TrackingError) Error() string {
	if te.Err == nil {
		return ""
	}
	return te.Err.Error()
}

// Unwrap returns the inner error.
func (te *TrackingError) Unwrap() error {
	return te.Err
}
