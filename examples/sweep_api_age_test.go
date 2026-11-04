// Copyright 2026, Pulumi Corporation.

//go:build api
// +build api

package examples

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSweepTimeParsesEachListFormat(t *testing.T) {
	want := time.Date(2026, 6, 24, 22, 46, 4, 55_000_000, time.UTC)
	for name, raw := range map[string]string{
		"rfc3339":      `"2026-06-24T22:46:04.055Z"`,
		"space-millis": `"2026-06-24 22:46:04.055"`,
		"epoch-millis": `1782341164055`,
	} {
		t.Run(name, func(t *testing.T) {
			var st sweepTime
			if err := json.Unmarshal([]byte(raw), &st); err != nil {
				t.Fatalf("unmarshal %s: %v", raw, err)
			}
			if !st.Equal(want) {
				t.Errorf("unmarshal %s = %v, want %v", raw, st.Time, want)
			}
		})
	}
}

func TestSweepTimeIsZeroWhenUnparseable(t *testing.T) {
	for _, raw := range []string{`{}`, `{"created":null}`, `{"created":"yesterday"}`, `{"created":0}`} {
		var resp struct {
			Created sweepTime `json:"created"`
		}
		if err := json.Unmarshal([]byte(raw), &resp); err != nil {
			t.Fatalf("unmarshal %s: %v", raw, err)
		}
		if !resp.Created.IsZero() {
			t.Errorf("unmarshal %s: created = %v, want zero", raw, resp.Created.Time)
		}
	}
}

func TestShouldSweep(t *testing.T) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	old := sweepTime{now.Add(-sweepMinAge - time.Minute)}
	recent := sweepTime{now.Add(-time.Minute)}
	const leaked = "test-env-12345"
	for _, tc := range []struct {
		name    string
		created sweepTime
		want    bool
	}{
		{leaked, old, true},
		{leaked, recent, false},
		{leaked, sweepTime{}, false},
		{"production", old, false},
	} {
		if got := shouldSweep(tc.name, tc.created, now); got != tc.want {
			t.Errorf("shouldSweep(%q, created %v before now) = %v, want %v",
				tc.name, now.Sub(tc.created.Time), got, tc.want)
		}
	}
}
