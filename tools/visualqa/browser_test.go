package main

import "testing"

func TestFullPageHeightGuard(t *testing.T) {
	tests := []struct {
		h    int
		want bool
	}{
		{0, true},
		{17679, true}, // proven live: Chrome 151 captured 1125x17679 fine
		{maxFullPageHeight, true},
		{maxFullPageHeight + 1, false},
	}
	for _, tt := range tests {
		if got := fullPageHeightOK(tt.h); got != tt.want {
			t.Errorf("fullPageHeightOK(%d) = %v, want %v", tt.h, got, tt.want)
		}
	}
}
