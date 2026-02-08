package dto

// UserExtendedResponse includes user info with devices and usage
type UserExtendedResponse struct {
	UserID           string                     `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email            string                     `json:"email" example:"user@example.com"`
	IP               string                     `json:"ip" example:"10.0.0.2"`
	PublicKey        string                     `json:"publicKey" example:"QGF4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y="`
	Status           string                     `json:"status" example:"active"`
	DataUsageLimit   int64                      `json:"dataUsageLimit" example:"107374182400"`
	CreatedAt        string                     `json:"createdAt" example:"2025-02-07T08:30:00Z"`
	UpdatedAt        string                     `json:"updatedAt" example:"2025-02-07T08:30:00Z"`
	LastConnected    *string                    `json:"lastConnected,omitempty" example:"2025-02-07T08:35:00Z"`
	DeviceCount      int                        `json:"deviceCount" example:"2"`
	ConnectedCount   int                        `json:"connectedCount" example:"1"`
	Devices          []UserDeviceInfo           `json:"devices"`
	UsageInfo        UserUsageInfo              `json:"usageInfo"`
}

// UserDeviceInfo is basic device info for user response
type UserDeviceInfo struct {
	DeviceID      string  `json:"deviceId" example:"a62557e6-4774-4e07-8ebb-9e89660df091"`
	DeviceName    string  `json:"deviceName" example:"My Phone"`
	DeviceType    string  `json:"deviceType" example:"mobile"`
	IP            string  `json:"ip" example:"192.168.89.2"`
	Status        string  `json:"status" example:"disconnected"`
	IsConnected   bool    `json:"isConnected" example:"false"`
	LastConnected *string `json:"lastConnected,omitempty" example:"2025-02-08T15:30:00Z"`
	CreatedAt     string  `json:"createdAt" example:"2025-02-08T14:00:00Z"`
}

// UserUsageInfo contains user's data usage information
type UserUsageInfo struct {
	TotalBytesSent     int64   `json:"totalBytesSent" example:"2147483648"`
	TotalBytesReceived int64   `json:"totalBytesReceived" example:"10737418240"`
	TotalBytes         int64   `json:"totalBytes" example:"12884901888"`
	TotalGB            float64 `json:"totalGB" example:"12.0"`
	LimitGB            float64 `json:"limitGB" example:"100.0"`
	UsedPercentage     float64 `json:"usedPercentage" example:"12.0"`
	RemainingGB        float64 `json:"remainingGB" example:"88.0"`
}
