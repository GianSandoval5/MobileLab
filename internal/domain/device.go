package domain

import (
	"context"
	"errors"
	"fmt"
)

type Capability string

const (
	CapabilityLaunch         Capability = "launch"
	CapabilityStop           Capability = "stop"
	CapabilityDeepLink       Capability = "deepLink"
	CapabilityLocation       Capability = "location"
	CapabilityNetworkOnline  Capability = "networkOnline"
	CapabilityNetworkOffline Capability = "networkOffline"
	CapabilityNetworkLatency Capability = "networkLatency"
	CapabilityPush           Capability = "push"
)

type CapabilityLevel string

const (
	CapabilityAvailable   CapabilityLevel = "available"
	CapabilityPartial     CapabilityLevel = "partial"
	CapabilityUnavailable CapabilityLevel = "unavailable"
)

type Device struct {
	ID           string                         `json:"id"`
	Name         string                         `json:"name"`
	Platform     string                         `json:"platform"`
	Emulator     bool                           `json:"emulator"`
	State        string                         `json:"state"`
	Details      map[string]string              `json:"details,omitempty"`
	Capabilities map[Capability]CapabilityLevel `json:"capabilities"`
}

type Location struct {
	Latitude  float64
	Longitude float64
}

type NetworkCondition string

const (
	NetworkOnline  NetworkCondition = "online"
	NetworkOffline NetworkCondition = "offline"
	NetworkSlow    NetworkCondition = "slow"
)

var ErrCapabilityUnavailable = errors.New("capability not available on this platform")

type DeviceAdapter interface {
	Platform() string
	Detect(context.Context) ([]Device, error)
	LaunchApp(context.Context, string, string) error
	StopApp(context.Context, string, string) error
	OpenDeepLink(context.Context, string, string) error
	SetLocation(context.Context, string, Location) error
	SetNetworkCondition(context.Context, string, NetworkCondition) error
}

type CapabilityError struct {
	Platform   string
	Capability Capability
	Reason     string
}

func (e CapabilityError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("%s: %s: %v", e.Platform, e.Capability, ErrCapabilityUnavailable)
	}
	return fmt.Sprintf("%s: %s: %v (%s)", e.Platform, e.Capability, ErrCapabilityUnavailable, e.Reason)
}

func (e CapabilityError) Unwrap() error { return ErrCapabilityUnavailable }
