/*
Copyright 2023 Adevinta
*/

package postgresql

import (
	"database/sql"
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/labstack/gommon/log"
)

var (
	baseModelFieldNames = []string{"ID"}
	ignoreFieldsTeam    = cmpopts.IgnoreFields(model.FindingTicket{}, baseModelFieldNames...)
)

func TestGetFindingTicket(t *testing.T) {
	testStore, err := PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)

	}
	defer testStore.Close()

	tests := []struct {
		name      string
		findingID string
		teamID    string
		want      model.FindingTicket
		wantErr   error
	}{
		{
			name:      "HappyPath",
			findingID: "89f84d26-9e5c-46ca-886f-9310357aba90",
			teamID:    "7b54fd77-91f3-46d5-8b08-828655bf076c",
			want: model.FindingTicket{
				ID:         "4bbf6ea1-ba1f-4edb-b6ee-59fa544af0cd",
				FindingID:  "89f84d26-9e5c-46ca-886f-9310357aba90",
				TeamID:     "7b54fd77-91f3-46d5-8b08-828655bf076c",
				URLTracker: "https://jira-server.com/browse/TEST-1",
			},
			wantErr: nil,
		},
		{
			name:      "FindingNotExists",
			findingID: "fd0c392e-80dc-4a0c-89b3-22ac0f0dbb81",
			teamID:    "7b54fd77-91f3-46d5-8b08-828655bf076c",
			want:      model.FindingTicket{},
			wantErr:   sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStore.GetFindingTicket(tt.findingID, tt.teamID)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Fatalf("finding ticket does not match expected one. diff: %s\n", diff)
			}
		})
	}
}

func TestCreateFindingTicket(t *testing.T) {
	testStore, err := PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
	if err != nil {
		log.Fatal(err)

	}

	defer testStore.Close()
	tests := []struct {
		name              string
		ticket            model.Ticket
		wantFindingTicket model.FindingTicket
		wantErr           error
	}{
		{
			name: "HappyPath",
			ticket: model.Ticket{
				TeamID:     "ee535e72-b514-4846-8eeb-9be175e00fdb",
				FindingID:  "acba85a5-89f2-4044-904c-a35b55fe15d5",
				URLTracker: "https://jira-server.com/browse/TEST-3",
			},
			wantFindingTicket: model.FindingTicket{
				TeamID:     "ee535e72-b514-4846-8eeb-9be175e00fdb",
				FindingID:  "acba85a5-89f2-4044-904c-a35b55fe15d5",
				URLTracker: "https://jira-server.com/browse/TEST-3",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := testStore.CreateFindingTicket(tt.ticket)
			if errToStr(err) != errToStr(tt.wantErr) {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
			diff := cmp.Diff(tt.wantFindingTicket, got, ignoreFieldsTeam)
			if diff != "" {
				t.Fatalf("ticket does not match expected one. diff: %s\n", diff)
			}
		})
	}
}
