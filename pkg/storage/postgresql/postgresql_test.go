package postgresql

import (
	"testing"

	"github.com/adevinta/vulcan-tracker/pkg/testutil"
	"github.com/google/go-cmp/cmp"
)

func errToStr(err error) string {
	return testutil.ErrToStr(err)
}

func TestHealthcheckOk(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "HealthcheckOK",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s, err := PrepareDatabaseLocal("../../../testdata/fixtures", NewDB)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()

			err = s.Healthcheck()

			diff := cmp.Diff(errToStr(err), errToStr(tt.wantErr))
			if diff != "" {
				t.Fatalf("expected error: %v but got: %v", tt.wantErr, err)
			}
		})
	}
}
