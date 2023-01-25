/*
Copyright 2022 Adevinta
*/

package testutils

func ErrToStr(err error) string {
	result := ""
	if err != nil {
		result = err.Error()
	}
	return result
}
