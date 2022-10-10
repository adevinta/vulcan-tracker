/*
Copyright 2022 Adevinta
*/

package jira

import (
	"fmt"

	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// GetTicket retrieves a ticket from Jira.
func (tc TC) GetTicket(id string) (*model.Ticket, error) {
	ticket, err := tc.Client.GetTicket(id)
	if err != nil {
		return nil, err
	}
	return ticket, nil

}

// CreateTicket creates an ticket in Jira.
func (tc TC) CreateTicket(ticket *model.Ticket) (*model.Ticket, error) {

	createdTicket, err := tc.Client.CreateTicket(ticket, ticket.TicketType)
	if err != nil {
		return nil, err
	}

	return createdTicket, nil
}

// GetTransitions retrieves all the transitions that are posibles from the curruent state of a ticket.
func (tc TC) GetTransitions(id string) ([]model.Transition, error) {
	transitions, err := tc.Client.GetTicketTransitions(id)
	if err != nil {
		return nil, err
	}

	return transitions, nil
}

// FixTicket transits a ticket until a "done" state.
// The states stored in workflow argument has to be in the order necessary to
// get a successful final state.
func (tc TC) FixTicket(id string, workflow []string) (*model.Ticket, error) {

	// First, check if the ticket exists.
	ticket, err := tc.Client.GetTicket(id)
	if err != nil {
		return nil, err
	}

	lastState := workflow[len(workflow)-1]

	//The ticket already has a "done" state.
	if ticket.Status == lastState {
		return ticket, nil
	}

	for _, transitionName := range workflow {

		// Get the available transitions for the ticket.
		transitions, err := tc.Client.GetTicketTransitions(id)
		if err != nil {
			return nil, err
		}

		for _, transition := range transitions {
			if transition.ToName == transitionName {
				err := tc.Client.DoTransition(id, transition.ID)
				if err != nil {
					return nil, err
				}
				break
			}
		}
	}

	ticket, err = tc.Client.GetTicket(id)
	if err != nil {
		return nil, err
	}

	// Check if the status of the ticket is successful.
	if ticket.Status != lastState {
		return nil, fmt.Errorf("it hasn't been possible to get a fixed status for the ticket %s", id)
	}

	return ticket, nil
}