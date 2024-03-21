/*
Copyright 2022 Adevinta
*/

// Package api contains the endpoints to manage finding tickets.
package api

import (
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
)

// API represents an API services and all the stuff needed to work.
type API struct {
	ticketServer         tracking.TicketServer
	ticketTrackerBuilder tracking.TicketTrackerBuilder
	storage              storage.Storage
	Options              Options
}

// Options represents size options for the API requests.
type Options struct {
	MaxSize     int
	DefaultSize int
}

// New instantiates a new API.
func New(ticketServer tracking.TicketServer, ticketTrackerBuilder tracking.TicketTrackerBuilder,
	storage storage.Storage, options Options) *API {
	return &API{
		ticketServer:         ticketServer,
		ticketTrackerBuilder: ticketTrackerBuilder,
		storage:              storage,
		Options:              options,
	}
}
