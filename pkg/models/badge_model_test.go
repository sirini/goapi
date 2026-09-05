package models

import "testing"

func TestIsSenstaClient(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{name: "android", key: CLIENT_SENSTA_ANDROID, want: true},
		{name: "ios", key: CLIENT_SENSTA_IOS, want: true},
		{name: "browser", key: "nubo-web", want: false},
		{name: "empty", key: "", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsSenstaClient(test.key); got != test.want {
				t.Fatalf("IsSenstaClient(%q) = %v, want %v", test.key, got, test.want)
			}
		})
	}
}
