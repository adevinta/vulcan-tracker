/*
Copyright 2022 Adevinta
*/

package jira

import (
	"fmt"
	"net/url"
	"path"

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
func (tc *TrackerClient) GetTicket(id string) (model.Ticket, error) {
	ticket, err := tc.Client.GetTicket(id)
	if err != nil {
		return model.Ticket{}, err
	}
	if ticket.URLTracker, err = url.JoinPath(tc.URL, path.Join("browse", ticket.Key)); err != nil {
		return model.Ticket{}, err
	}

	return ticket, nil
}

// CreateTicket creates a ticket in Jira.
func (tc *TrackerClient) CreateTicket(ticket model.Ticket) (model.Ticket, error) {
	ticket.Description = GenerateDescriptionText(ticket.Description, ticket.FindingID, ticket.TeamID)

	createdTicket, err := tc.Client.CreateTicket(ticket)
	if err != nil {
		return model.Ticket{}, err
	}

	if createdTicket.URLTracker, err = url.JoinPath(tc.URL, path.Join("browse", createdTicket.Key)); err != nil {
		return model.Ticket{}, err
	}

	return createdTicket, nil
}
