/*
Copyright 2023 Adevinta
*/
package common

import "fmt"

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
