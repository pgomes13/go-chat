package commons

import "time"

const (
	// WebSocket timing
	WriteWait      = 10 * time.Second
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 512

	// MongoDB
	CollName = "messages"

	// Auth
	SessionName = "go-chat"
)
