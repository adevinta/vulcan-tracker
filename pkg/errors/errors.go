/*
Copyright 2022 Adevinta
*/
package errors

type TrackingError struct {
	Err            error
	HttpStatusCode int
}

func (te *TrackingError) Error() string {
	return te.Err.Error()
}

func (te *TrackingError) Unwrap() error {
	return te.Err
}
