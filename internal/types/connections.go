package types

import "time"

// ClashConnection represents a single connection from the Clash API
type ClashConnection struct {
	ID             string                 `json:"id"`
	Metadata       ConnectionMetadata     `json:"metadata"`
	Upload         int64                  `json:"upload"`
	Download       int64                  `json:"download"`
	Start          time.Time              `json:"start"`
	Chains         []string               `json:"chains"`
	Rule           string                 `json:"rule"`
	RulePayload    string                 `json:"rulePayload"`
}

// ConnectionMetadata contains metadata about a connection
type ConnectionMetadata struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	DestinationIP   string `json:"destinationIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
	DNSMode         string `json:"dnsMode"`
	ProcessPath     string `json:"processPath"`
}

// ClashConnectionsMessage represents the WebSocket message from Clash API
type ClashConnectionsMessage struct {
	Connections   []ClashConnection `json:"connections"`
	DownloadTotal int64             `json:"downloadTotal"`
	UploadTotal   int64             `json:"uploadTotal"`
	Memory        int64             `json:"memory"`
}

// ConnectionFilter represents filter criteria for connections
type ConnectionFilter struct {
	Network         string
	SourceIP        string
	DestinationIP   string
	DestinationPort string
	Host            string
	Chain           string
	Rule            string
}
