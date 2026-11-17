package analytics

import (
	"testing"
	"time"
)

func TestOverviewStartUsesBusinessMidnightRegardlessOfHostTimezone(t *testing.T) {
	for _, test := range []struct {
		now  string
		days uint
		want string
	}{
		{"2026-11-17T04:00:00Z", 1, "2026-11-17T00:00:00+08:00"},
		{"2026-11-16T17:00:00Z", 1, "2026-11-17T00:00:00+08:00"},
		{"2026-11-17T04:00:00Z", 30, "2026-10-19T00:00:00+08:00"},
	} {
		now, err := time.Parse(time.RFC3339, test.now)
		if err != nil {
			t.Fatal(err)
		}
		if got := overviewStart(now, test.days).Format(time.RFC3339); got != test.want {
			t.Fatalf("now=%s days=%d: got %s want %s", test.now, test.days, got, test.want)
		}
	}
}
