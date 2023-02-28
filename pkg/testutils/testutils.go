/*
Copyright 2022 Adevinta
*/

package testutils

// ErrToStr returns a string even when it is nil.
func ErrToStr(err error) string {
	result := ""
	if err != nil {
		result = err.Error()
	}
	return result
}
