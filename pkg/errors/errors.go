/*
Copyright 2022 Adevinta
*/
package errors

// TrackingError represents a custom error for this project.
type TrackingError struct {
	Err            error
	HTTPStatusCode int
	Msg            string
}

// Error returns an error as string.
func (te *TrackingError) Error() string {
	return te.Msg
}

// Unwrap returns the inner error.
func (te *TrackingError) Unwrap() error {
	return te.Err
}
