/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"context"
	"database/sql"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/lib/pq"
)

type FindingTicket struct {
	model.FindingTicket
	CreatedAt pq.NullTime `db:"created_at"`
	UpdatedAt pq.NullTime `db:"updated_at"`
}

// CreateFindingTicket inserts a row in the database to store the relation between a finding, a team and a ticket.
func (db DB) CreateFindingTicket(t model.Ticket) (model.FindingTicket, error) {
	tx, err := db.DBRw.BeginTxx(context.Background(), nil)
	if err != nil {
		return model.FindingTicket{}, err
	}

	query := `
	WITH ticket_exists AS (
		SELECT *
		FROM finding_tickets t
		WHERE t.finding_id = $1 and t.team_id = $2
	),
	creation_query AS (
	    INSERT INTO finding_tickets (finding_id, team_id, url_tracker, created_at)
			SELECT $1, $2, $3, now()
    	WHERE NOT EXISTS (SELECT 1 FROM ticket_exists)
	    RETURNING *	    
	)
	SELECT * FROM ticket_exists
    	UNION ALL
	SELECT * FROM creation_query`

	r := tx.QueryRowContext(context.Background(), query, t.FindingID, t.TeamID, t.URLTracker)
	if err != nil {
		tx.Rollback()
		return model.FindingTicket{}, err
	}
	var created FindingTicket
	err = r.Scan(&created.ID, &created.FindingID, &created.TeamID, &created.URLTracker, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return model.FindingTicket{}, err
	}
	return model.FindingTicket{
		ID:         created.ID,
		FindingID:  created.FindingID,
		TeamID:     created.TeamID,
		URLTracker: created.URLTracker,
	}, tx.Commit()
}

// GetFindingTicket retrieves a row from the database that contains the link to the ticket tracker for a specific
// finding and team.
func (db DB) GetFindingTicket(findingID, teamID string) (model.FindingTicket, error) {
	var findingTicket FindingTicket

	query := "SELECT * FROM finding_tickets WHERE finding_id = $1 and team_id = $2"
	logQuery(db.Logger, "GetFindingTicket", query, findingID, teamID)
	err := db.DB.Get(&findingTicket, query, findingID, teamID)
	if err == sql.ErrNoRows {
		return model.FindingTicket{}, nil
	}
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
