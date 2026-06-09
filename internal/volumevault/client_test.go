package volumevault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetVolumesSupportsWrappedAndDirectResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "wrapped",
			body: `{"data":[{"id":1,"name":"vol1","backup_state":"backed_up"}]}`,
		},
		{
			name: "direct",
			body: `[{"id":1,"name":"vol1","backup_state":"backed_up"}]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/volumes" {
					t.Fatalf("unexpected path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer ts.Close()

			c := NewClient(ts.URL, "token", 2*time.Second)
			vols, err := c.GetVolumes(context.Background())
			if err != nil {
				t.Fatalf("GetVolumes() error = %v", err)
			}
			if len(vols) != 1 {
				t.Fatalf("len(vols) = %d, want 1", len(vols))
			}
			if vols[0].Name != "vol1" {
				t.Fatalf("vols[0].Name = %q, want vol1", vols[0].Name)
			}
		})
	}
}
