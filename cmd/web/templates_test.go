package main

import (
	"testing"
	"time"

	"snippetbox.mabona3.net/internal/assert"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2025, 5, 17, 10, 15, 0, 0, time.UTC),
			want: "17 May 2025 at 10:15",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2025, 5, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 May 2025 at 09:15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			assert.Equal(t, humanDate(tt.tm), tt.want)
		})
	}
}
