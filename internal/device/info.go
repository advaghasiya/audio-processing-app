package device

import (
	"net/http"

	"github.com/mssola/user_agent"
)

type DeviceInfo struct {
	Browser        string
	BrowserVersion string
	OS             string
	OSVersion      string
	Device         string
	IsMobile       bool
	IsTablet       bool
	IsPC           bool
	IPAddress      string
}

func GetDeviceInfo(r *http.Request) *DeviceInfo {
	ua := user_agent.New(r.UserAgent())
	browser, browserVersion := ua.Browser()

	info := &DeviceInfo{
		Browser:        browser,
		BrowserVersion: browserVersion,
		OS:             ua.OS(),
		IsMobile:       ua.Mobile(),
		IsTablet:       ua.Mobile(),
		IPAddress:      r.RemoteAddr,
	}

	if !info.IsMobile && !info.IsTablet {
		info.IsPC = true
	}

	info.Device = ua.Platform()

	osInfo := ua.OSInfo()
	if osInfo.Name != "" {
		info.OS = osInfo.Name
		info.OSVersion = osInfo.Version
	}

	return info
}
