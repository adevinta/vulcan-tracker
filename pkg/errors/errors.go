/*
Copyright 2022 Adevinta
*/
package errors

type TrackingError struct {
	Err            error
	HttpStatusCode int
	Msg            string
}

func (te *TrackingError) Error() string {
	return te.Msg
}

func (te *TrackingError) Unwrap() error {
	return te.Err
}
