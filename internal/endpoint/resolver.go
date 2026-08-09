package endpoint

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/domain"
)

type Result struct {
	URL      string `json:"url"`
	Platform string `json:"platform"`
	DeviceID string `json:"device_id,omitempty"`
	Note     string `json:"note,omitempty"`
}

func Resolve(configuration config.Config, platform string, target *domain.Device) (Result, error) {
	if err := configuration.Validate(); err != nil {
		return Result{}, fmt.Errorf("invalid MobileLab configuration: %w", err)
	}
	if target != nil {
		if platform != "" && target.Platform != platform {
			return Result{}, fmt.Errorf("device %q belongs to %s, not %s", target.ID, target.Platform, platform)
		}
		platform = target.Platform
	}
	host := configuration.Server.Host
	result := Result{Platform: platform}
	if target != nil {
		result.DeviceID = target.ID
	}

	switch platform {
	case "":
		host = localHost(host)
		result.Platform = "host"
	case "android":
		if target == nil {
			return Result{}, fmt.Errorf("an Android device is required to resolve its endpoint")
		}
		if isLoopbackHost(host) {
			if !target.Emulator {
				return Result{}, fmt.Errorf("Android device %q cannot reach host loopback directly; configure a trusted-network server.host or run adb reverse tcp:%d tcp:%d", target.ID, configuration.Server.Port, configuration.Server.Port)
			}
			host = "10.0.2.2"
			result.Note = "Standard Android Emulator host alias; other emulator products may use a different address."
		} else if isUnspecifiedHost(host) {
			return Result{}, fmt.Errorf("server.host %q is a bind address, not a reachable Android address; configure a specific trusted-network host", host)
		}
	case "ios":
		if target == nil {
			return Result{}, fmt.Errorf("an iOS device is required to resolve its endpoint")
		}
		if isLoopbackHost(host) {
			if !target.Emulator {
				return Result{}, fmt.Errorf("iOS device %q cannot reach host loopback directly; configure a specific trusted-network server.host", target.ID)
			}
			host = localHost(host)
			result.Note = "iOS Simulator shares the Mac host network namespace."
		} else if isUnspecifiedHost(host) {
			return Result{}, fmt.Errorf("server.host %q is a bind address, not a reachable iOS address; configure a specific trusted-network host", host)
		}
	default:
		return Result{}, fmt.Errorf("platform must be android or ios")
	}

	result.URL = (&url.URL{Scheme: "http", Host: net.JoinHostPort(host, strconv.Itoa(configuration.Server.Port))}).String()
	return result, nil
}

func localHost(host string) string {
	if host == "localhost" {
		return host
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.To4() == nil && (ip.IsLoopback() || ip.IsUnspecified()) {
		return "::1"
	}
	if ip != nil && (ip.IsLoopback() || ip.IsUnspecified()) {
		return "127.0.0.1"
	}
	return host
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func isUnspecifiedHost(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && ip.IsUnspecified()
}
