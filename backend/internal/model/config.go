package model

type DeviceType string

const (
	DeviceWeb    DeviceType = "web"
	DevicePC     DeviceType = "pc"
	DeviceMobile DeviceType = "mobile"
)

type ClientConfig struct {
	Device  DeviceType       `json:"device"`
	Configs map[string]any   `json:"configs"`
}
