package main

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
)

// deviceProfile describes one emulated form factor. The numeric values are
// pinned by TestDevicePresets (rod v0.114.8: devices.IPhoneX 375x812@3,
// devices.IPad 768x1024@2; desktop uses a plain viewport override).
type deviceProfile struct {
	Name     string
	Viewport [2]int
	DPR      float64
	Touch    bool
	Mobile   bool
	UA       string
	emulate  bool
	preset   devices.Device
}

var devicePresets = map[string]deviceProfile{
	"mobile": {
		Name:     "mobile",
		Viewport: [2]int{375, 812},
		DPR:      3,
		Touch:    true,
		Mobile:   true,
		UA:       "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
		emulate:  true,
		preset:   devices.IPhoneX,
	},
	"tablet": {
		Name:     "tablet",
		Viewport: [2]int{768, 1024},
		DPR:      2,
		Touch:    true,
		Mobile:   true,
		UA:       "Mozilla/5.0 (iPad; CPU OS 11_0 like Mac OS X) AppleWebKit/604.1.34 (KHTML, like Gecko) Version/11.0 Mobile/15A5341f Safari/604.1",
		emulate:  true,
		preset:   devices.IPad,
	},
	"desktop": {
		Name:     "desktop",
		Viewport: [2]int{1440, 900},
		DPR:      1,
		Touch:    false,
		Mobile:   false,
		UA:       "",
		emulate:  false,
	},
}

// deviceProfileFor returns the profile for name, or ok=false when unknown.
func deviceProfileFor(name string) (deviceProfile, bool) {
	p, ok := devicePresets[name]
	return p, ok
}

// apply configures a rod page with this profile's emulation.
func (d deviceProfile) apply(page *rod.Page) {
	if d.emulate {
		page.MustEmulate(d.preset)
		return
	}
	page.MustSetViewport(d.Viewport[0], d.Viewport[1], d.DPR, false)
}
