/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/lib/pq"
)

// FindingTicket represents
type FindingTicket struct {
	model.FindingTicket
	CreatedAt pq.NullTime `db:"created_at"`
	UpdatedAt pq.NullTime `db:"updated_at"`
}

// CreateFindingTicket inserts a row in the database to store the relation between a finding, a team and a ticket.
func (db DB) CreateFindingTicket(t model.Ticket) (model.FindingTicket, error) {
	query := `
	    INSERT INTO finding_tickets (finding_id, team_id, url_tracker, created_at)
			SELECT $1, $2, $3, now() RETURNING *
			`
	result := db.DB.QueryRow(query, t.FindingID, t.TeamID, t.URLTracker)

	var created FindingTicket
	err := result.Scan(&created.ID, &created.FindingID, &created.TeamID, &created.URLTracker, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return model.FindingTicket{}, err
	}
	return model.FindingTicket{
		ID:         created.ID,
		FindingID:  created.FindingID,
		TeamID:     created.TeamID,
		URLTracker: created.URLTracker,
	}, nil
}

// GetFindingTicket retrieves a row from the database that contains the link to the ticket tracker for a specific
// finding and team.
func (db DB) GetFindingTicket(findingID, teamID string) (model.FindingTicket, error) {
	var findingTicket FindingTicket

	query := "SELECT * FROM finding_tickets WHERE finding_id = $1 and team_id = $2"
	logQuery(db.Logger, "GetFindingTicket", query, findingID, teamID)
	err := db.DB.Get(&findingTicket, query, findingID, teamID)
	if err != nil {
		return model.FindingTicket{}, err
	}

	return model.FindingTicket{
		ID:         findingTicket.ID,
		FindingID:  findingTicket.FindingID,
		TeamID:     findingTicket.TeamID,
		URLTracker: findingTicket.URLTracker,
	}, nil
}
