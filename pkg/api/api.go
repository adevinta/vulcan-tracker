/*
Copyright 2022 Adevinta
*/

package api

import (
	"errors"

	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
)

// API represents an API services and all the stuff needed to work.
type API struct {
	ticketServerStorage  storage.TicketServerStorage
	ticketTrackerBuilder tracking.TicketTrackerBuilder
	storage              storage.Storage
	Options              Options
}

// Options represents size options for the API requests.
type Options struct {
	MaxSize     int
	DefaultSize int
}

// Pagination represents the pagination options for the API requests.
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
	ErrDateMalformed = errors.New("malformed Date")

	// ErrPageMalformed indicates that the page requested is not an integer larger than 0.
	ErrPageMalformed = errors.New("malformed Page Number")

	// ErrPageNotFound indicates that the page requested does not exist.
	ErrPageNotFound = errors.New("page Not Found")

	// ErrSizeMalformed indicates that the size requested is not an integer larger than 0.
	ErrSizeMalformed = errors.New("malformed Size Number")

	// ErrSizeTooLarge indicates that the size requested is larger than the maximum allowed.
	ErrSizeTooLarge = errors.New("size Number Too Large")

	// ErrInvalidFilter indicates that there is a conflict between specified params for the filter.
	ErrInvalidFilter = errors.New("filter parameters combination is invalid")
)

// New instantiates a new API.
func New(ticketServerStorage storage.TicketServerStorage, ticketTrackerBuilder tracking.TicketTrackerBuilder,
	storage storage.Storage, options Options) *API {
	return &API{
		ticketServerStorage:  ticketServerStorage,
		ticketTrackerBuilder: ticketTrackerBuilder,
		storage:              storage,
		Options:              options,
	}
}
