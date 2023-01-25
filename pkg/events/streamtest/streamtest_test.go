/*
Copyright 2022 Adevinta
*/
package streamtest

import (
	"testing"
	"unicode"

	"github.com/adevinta/vulcan-tracker/pkg/events"
	"github.com/google/go-cmp/cmp"
)

func removeSpaces(s string) string {
	rr := make([]rune, 0, len(s))
	for _, r := range s {
		if !unicode.IsSpace(r) {
			rr = append(rr, r)
		}
	}
	return string(rr)
}

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		want        []events.Message
		shouldPanic bool
	}{
		{
			name:     "valid file",
			filename: "testdata/valid.json",
			want: []events.Message{
				{
					Key: []byte("FindingID-1"),
					Value: []byte(removeSpaces(`
					{
						"id": "FindingID-1",
						"affected_resource": "AffectedResource-1",
						"score": 9,
						"status": "OPEN",
						"details": "Details-1",
						"impact_details": "ImpactDetails-1",
						"issue": {
							"id": "IssueID-1",
							"summary": "Summary-1",
							"cwe_id": 1,
							"description": "Description-1",
							"recommendations": [
								"Recommendation-1",
								"Recommendation-2"
							],
							"reference_links": [
								"ReferenceLink-1",
								"ReferenceLink-2"
							],
							"labels": [
								"Label-1",
								"Label-2"
							]
						},
						"target": {
							"id": "TargetID-1",
							"identifier": "Identifier-1",
							"teams": [
								"Team-1",
								"Team-2"
							]
						},
						"source": {
							"id": "SourceID-1",
							"instance": "SourceInstance-1",
							"options": "SourceOptions-1",
							"time": "0001-01-01T00:00:00Z",
							"name": "SourceName-1",
							"component": "SourceComponent-1"
						},
						"resources": [
							{
								"name": "ResourceName-1",
								"attributes": [
									"Attr-1",
									"Attr-2"
								],
								"resources": [
									{
										"Attr-1": "1",
										"Attr-2": "2"
									}
								]
							}
						],
						"total_exposure": 10,
						"current_exposure": 5
					}
					`)),
					Headers: []events.MetadataEntry{
						{
							Key:   []byte("version"),
							Value: []byte("0.0.1"),
						},
					},
				},
			},
			shouldPanic: false,
		},
		{
			name:        "malformed json",
			filename:    "testdata/malformed_json.json",
			want:        nil,
			shouldPanic: true,
		},
		{
			name:        "null metadata key",
			filename:    "testdata/malformed_null_metadata_key.json",
			want:        nil,
			shouldPanic: true,
		},
		{
			name:        "null metadata value",
			filename:    "testdata/malformed_null_metadata_value.json",
			want:        nil,
			shouldPanic: true,
		},
		{
			name:        "nonexistent file",
			filename:    "testdata/nonexistent.json",
			want:        nil,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); (err != nil) != tt.shouldPanic {
					t.Errorf("unexpected panic behavior: %v", err)
				}
			}()

			got := MustParse(tt.filename)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("asset mismatch (-want +got):\n%v", diff)
			}
		})
	}
}
