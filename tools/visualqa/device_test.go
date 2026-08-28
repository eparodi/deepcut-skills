package main

import "testing"

func TestDevicePresets(t *testing.T) {
	tests := []struct {
		name     string
		viewport [2]int
		dpr      float64
		touch    bool
		mobile   bool
		ua       string
	}{
		{name: "mobile", viewport: [2]int{375, 812}, dpr: 3, touch: true, mobile: true, ua: "iPhone"},
		{name: "tablet", viewport: [2]int{768, 1024}, dpr: 2, touch: true, mobile: true, ua: "iPad"},
		{name: "desktop", viewport: [2]int{1440, 900}, dpr: 1, touch: false, mobile: false, ua: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := deviceProfileFor(tt.name)
			if !ok {
				t.Fatalf("deviceProfileFor(%q) = ok=false, want true", tt.name)
			}
			if p.Viewport != tt.viewport {
				t.Errorf("Viewport = %v, want %v", p.Viewport, tt.viewport)
			}
			if p.DPR != tt.dpr {
				t.Errorf("DPR = %v, want %v", p.DPR, tt.dpr)
			}
			if p.Touch != tt.touch {
				t.Errorf("Touch = %v, want %v", p.Touch, tt.touch)
			}
			if p.Mobile != tt.mobile {
				t.Errorf("Mobile = %v, want %v", p.Mobile, tt.mobile)
			}
			if tt.ua != "" && !contains(p.UA, tt.ua) {
				t.Errorf("UA = %q, want it to contain %q", p.UA, tt.ua)
			}
			if tt.ua == "" && p.UA != "" {
				t.Errorf("UA = %q, want empty for desktop", p.UA)
			}
		})
	}
}

func TestDeviceProfileForUnknown(t *testing.T) {
	if _, ok := deviceProfileFor("smartwatch"); ok {
		t.Fatal("deviceProfileFor(smartwatch) = ok=true, want false")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
