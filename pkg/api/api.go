/*
Copyright 2022 Adevinta
*/

package api

import (
	"errors"

	"github.com/adevinta/vulcan-jira-api/pkg/issues"
)

type API struct {
	issueTracking issues.IssueTracking
	Options       Options
}

type Options struct {
	MaxSize     int
	DefaultSize int
}

type Pagination struct {
	Limit  int  `json:"limit"`
	Offset int  `json:"offset"`
	Total  int  `json:"total"`
	More   bool `json:"more"`
}

var (
	// Regular expression matching date format 'yyyy-mm-dd'.
	dateFmtRegEx = `^\d{4}\-(0[1-9]|1[012])\-(0[1-9]|[12][0-9]|3[01])$`

	// ErrDateMalformed indicates that the date format does not comply with YYYY-MM-DD.
	ErrDateMalformed = errors.New("Malformed Date")

	// ErrPageMalformed indicates that the page requested is not an integer larger than 0.
	ErrPageMalformed = errors.New("Malformed Page Number")

	// ErrPageNotFound indicates that the page requested does not exist.
	ErrPageNotFound = errors.New("Page Not Found")

	// ErrSizeMalformed indicates that the size requested is not an integer larger than 0.
	ErrSizeMalformed = errors.New("Malformed Size Number")

	// ErrSizeTooLarge indicates that the size requested is larger than the maximum allowed.
	ErrSizeTooLarge = errors.New("Size Number Too Large")

	// ErrInvalidFilter indicates that there is a conflict between specified params for the filter.
	ErrInvalidFilter = errors.New("Filter parameters combination is invalid")
)

// New instantiates a new API.
func New(issueTracking issues.IssueTracking, options Options) *API {
	return &API{
		issueTracking: issueTracking,
		Options:       options,
	}
}
