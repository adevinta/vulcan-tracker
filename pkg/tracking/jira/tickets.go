/*
Copyright 2022 Adevinta
*/
package jira

import (
	"fmt"

	"github.com/adevinta/vulcan-tracker/pkg/model"
)

// GenerateIdentificationText generate the part of the description that contains the identifiers.
func GenerateIdentificationText(findingID, teamID string) string {
	return fmt.Sprintf("FindingID: %s\nTeamID: %s", findingID, teamID)
}

// GenerateDescriptionText generate description using the original description and the finding and team identifiers.
func GenerateDescriptionText(ticketDescription, findingID, teamID string) string {
	beginAutomaticText := "======= BEGINNING OF THE CONTENT AUTOMATICALLY INSERTED ======\n"
	dontRemoveText := "======= PLEASE, DON'T REMOVE THE TEXT BETWEEN THESE MARKS =====\n"
	endAutomaticText := "======= END OF THE CONTENT AUTOMATICALLY INSERTED ============\n"
	ticketIdentificationText := GenerateIdentificationText(findingID, teamID)
	return fmt.Sprintf("\n%s\n\n%s%s%s\n%s",
		ticketDescription, beginAutomaticText, dontRemoveText, ticketIdentificationText, endAutomaticText)
}

// GetTicket retrieves a ticket from Jira.
// Return an empty ticket if not found.
func (tc *TC) GetTicket(id string) (model.Ticket, error) {
	ticket, err := tc.Client.GetTicket(id)
	if err != nil {
		return model.Ticket{}, err
	}
	ticket.URLTracker = fmt.Sprintf("%s/browse/%s", tc.URL, ticket.Key)

	return ticket, nil
}

// FindTicketByFindingAndTeam retrieves a ticket from Jira using the project key, the vulnerability issue type,
// the finding ID and the team ID.
// Return an empty ticket if not found.
func (tc *TC) FindTicketByFindingAndTeam(projectKey, vulnerabilityIssueType, findingID string, teamID string) (model.Ticket, error) {
	text := fmt.Sprintf("\"%s\"", GenerateIdentificationText(findingID, teamID))

	ticket, err := tc.Client.FindTicket(projectKey, vulnerabilityIssueType, text)
	if err != nil {
		return model.Ticket{}, err
	}
	return ticket, nil
}

// CreateTicket creates a ticket in Jira.
func (tc *TC) CreateTicket(ticket model.Ticket) (model.Ticket, error) {

	ticket.Description = GenerateDescriptionText(ticket.Description, ticket.FindingID, ticket.TeamID)

	createdTicket, err := tc.Client.CreateTicket(ticket)
	if err != nil {
		return model.Ticket{}, err
	}

	createdTicket.URLTracker = fmt.Sprintf("%s/browse/%s", tc.URL, createdTicket.Key)

	return createdTicket, nil
}

// GetTransitions retrieves all the transitions that are posibles from the curruent state of a ticket.
func (tc *TC) GetTransitions(id string) ([]model.Transition, error) {
	transitions, err := tc.Client.GetTicketTransitions(id)
	if err != nil {
		return nil, err
	}

	return transitions, nil
}

// getSubWorkflow returns a subslice starting by state.
func getSubWorkflow(workflow []string, state string) []string {
	for i, wState := range workflow {
		if wState == state {
			return workflow[i:]
		}
	}
	return []string{}

}

// FixTicket transits a ticket until a "done" state.
// The states stored in workflow argument has to be in the order necessary to
// get a successful final state.
// Return an empty ticket if not found.
func (tc *TC) FixTicket(id string, workflow []string) (model.Ticket, error) {
	// First, check if the ticket exists.
	ticket, err := tc.Client.GetTicket(id)
	if err != nil {
		return model.Ticket{}, err
	}

	if ticket.ID == "" {
		return ticket, nil
	}

	lastState := workflow[len(workflow)-1]

	//The ticket already has a "done" state.
	if ticket.Status == lastState {
		return ticket, nil
	}

	//Start the workflow from the current ticket status.
	if len(workflow) > 1 {
		workflow = getSubWorkflow(workflow, ticket.Status)
	}

	// TODO: When we can to transit from any state to a fixed state we need a list of valid states to do this.
	// For example. If the ticket is in a CLOSED state we don't want acidentally go back to RESOLVED.

	for _, transitionName := range workflow {

		// Get the available transitions for the ticket.
		transitions, err := tc.Client.GetTicketTransitions(id)
		if err != nil {
			return model.Ticket{}, err
		}

		for _, transition := range transitions {
			if transition.ToName == transitionName {
				var err error
				if transitionName == lastState {
					err = tc.Client.DoTransitionWithResolution(id, transition.ID, "Done")
				} else {
					err = tc.Client.DoTransition(id, transition.ID)
				}
				if err != nil {
					return model.Ticket{}, err
				}
				break
			}
		}
	}

	ticket, err = tc.Client.GetTicket(id)
	if err != nil {
		return model.Ticket{}, err
	}

	// Check if the status of the ticket is successful.
	if ticket.Status != lastState {
		return model.Ticket{}, fmt.Errorf("it hasn't been possible to get a fixed status for the ticket %s", id)
	}

	return ticket, nil
}

// WontFixTicket transits a ticket until a "done" state.
// The states stored in workflow argument has to be in the order necessary to
// get a successful final state.
// Return an empty ticket if not found.
func (tc *TC) WontFixTicket(id string, workflow []string, reason string) (model.Ticket, error) {
	// First, check if the ticket exists.
	ticket, err := tc.Client.GetTicket(id)
	if err != nil {
		return model.Ticket{}, err
	}

	if ticket.ID == "" {
		return ticket, nil
	}

	lastState := workflow[len(workflow)-1]

	//The ticket already has a "won't fix" state.
	if ticket.Status == lastState {
		return ticket, nil
	}

	//Start the workflow from the current ticket status.
	if len(workflow) > 1 {
		workflow = getSubWorkflow(workflow, ticket.Status)
	}

	for _, transitionName := range workflow {

		// Get the available transitions for the ticket.
		transitions, err := tc.Client.GetTicketTransitions(id)
		if err != nil {
			return model.Ticket{}, err
		}

		for _, transition := range transitions {
			if transition.ToName == transitionName {
				var err error
				if transitionName == lastState {
					err = tc.Client.DoTransitionWithResolution(id, transition.ID, reason)
				} else {
					err = tc.Client.DoTransition(id, transition.ID)
				}
				if err != nil {
					return model.Ticket{}, err
				}
				break
			}
		}
	}

	ticket, err = tc.Client.GetTicket(id)
	if err != nil {
		return model.Ticket{}, err
	}

	// Check if the status of the ticket is successful.
	if ticket.Status != lastState {
		return model.Ticket{}, fmt.Errorf("it hasn't been possible to get a won't fix status for the ticket %s", id)
	}

	return ticket, nil
}
