package repositories

import (
	"testing"
	"time"
)

func TestVerificationCodeValid(t *testing.T) {
	now := time.UnixMilli(2_000_000)
	tests := []struct {
		name      string
		timestamp int64
		want      bool
	}{
		{name: "fresh", timestamp: now.Add(-time.Minute).UnixMilli(), want: true},
		{name: "boundary", timestamp: now.Add(-verificationCodeLifetime).UnixMilli(), want: true},
		{name: "expired", timestamp: now.Add(-verificationCodeLifetime - time.Millisecond).UnixMilli(), want: false},
		{name: "future", timestamp: now.Add(time.Millisecond).UnixMilli(), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verificationCodeValid(tt.timestamp, now); got != tt.want {
				t.Fatalf("verificationCodeValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
