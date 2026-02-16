package model

// WiFiConnectionInfo represents the output from termux-wifi-connectioninfo command
type WiFiConnectionInfo struct {
	BSSID           string `json:"bssid"`
	FrequencyMhz    int    `json:"frequency_mhz"`
	IP              string `json:"ip"`
	LinkSpeedMbps   int    `json:"link_speed_mbps"`
	MACAddress      string `json:"mac_address"`
	NetworkID       int    `json:"network_id"`
	RSSI            int    `json:"rssi"`
	SSID            string `json:"ssid"`
	SSIDHidden      bool   `json:"ssid_hidden"`
	SupplicantState string `json:"supplicant_state"`
}
